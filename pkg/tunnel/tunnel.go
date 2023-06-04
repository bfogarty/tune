package tunnel

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

type Tunnel struct {
	LocalPort  int
	RemotePort int
	RemoteHost string
	AwsRegion  string

	KeepAliveInterval time.Duration

	ssmCmd *exec.Cmd
	target *ssh.Client
}

func New(remoteHost string, localPort int, remotePort int, awsRegion string) (*Tunnel, error) {
	return &Tunnel{
		LocalPort:  localPort,
		RemotePort: remotePort,
		RemoteHost: remoteHost,
		AwsRegion:  awsRegion,

		KeepAliveInterval: 10 * time.Second,
	}, nil
}

func (t *Tunnel) Start() error {
	// begin listening on localhost:port
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", t.LocalPort))
	if err != nil {
		return err
	}

	// intercept Ctrl+C
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		log.Printf("Shutting down...")
		listener.Close()
	}()

	// establish the connection to the target instance
	err = t.dial()
	if err != nil {
		return err
	}
	defer t.target.Close()
	defer t.ssmCmd.Process.Kill()

	// start the keep-alive routine
	go t.keepAlive()

	// handle connections
	log.Printf("Listening on localhost:%d", t.LocalPort)
	for {
		client, err := listener.Accept()
		if err != nil {
			return nil
		}
		log.Printf("Accepting connection...")
		go t.forward(client)
	}
}

func (t *Tunnel) dial() error {
	// get a random jump instance
	instance, err := getJumpInstance(t.AwsRegion)
	if err != nil {
		return err
	}
	log.Printf("Found jump instance: %s", instance.ID)

	// create an ephemeral SSH key valid for 60s
	log.Printf("Generating ephemeral ED25519 key...")
	privateKey, publicKey, err := generateKeyPair()
	if err != nil {
		return err
	}

	log.Printf("Sending ephemeral RSA key to %s...", instance.ID)
	err = sendKey(publicKey, instance.ID, instance.AvailabilityZone, t.AwsRegion)
	if err != nil {
		return err
	}

	// create the SSH config
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return err
	}
	sshConfig := &ssh.ClientConfig{
		User:            "ec2-user",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// use a pipe to redirect stdin/stdout to the ProxyCommand
	// TODO defer client.Close(), server.Close()
	client, server := net.Pipe()

	// start the SSH session over SSM
	t.ssmCmd = exec.Command("aws", []string{
		"ssm",
		"start-session",
		"--target",
		instance.ID,
		"--document-name",
		"AWS-StartSSHSession",
		"--parameters",
		"portNumber=22",
	}...)

	// make SSM read and write to the server side of the pipe
	t.ssmCmd.Stdin = server
	t.ssmCmd.Stdout = server
	t.ssmCmd.Stderr = os.Stderr

	log.Printf("Starting SSH session via SSM...")
	if err := t.ssmCmd.Start(); err != nil {
		return err
	}

	// make SSH use the client side of the pipe as its transport
	log.Printf("Connecting to %s...", instance.ID)
	host := fmt.Sprintf("%s:%d", instance.ID, 22)
	conn, chans, reqs, err := ssh.NewClientConn(client, host, sshConfig)
	if err != nil {
		return err
	}
	t.target = ssh.NewClient(conn, chans, reqs)

	return nil
}

func (t *Tunnel) keepAlive() {
	ticker := time.NewTicker(t.KeepAliveInterval)
	for range ticker.C {
		_, _, err := t.target.SendRequest("keepalive", true, nil)
		if err != nil {
			log.Printf("Error sending keep-alive request: %v", err)
		}
	}
}

func (t *Tunnel) forward(local net.Conn) error {
	defer func() {
		log.Printf("Closing local connection...")
		local.Close()
	}()
	done := make(chan bool, 1)

	// establish a connection to the remote
	log.Printf("Dialing remote: %s:%d...", t.RemoteHost, t.RemotePort)
	remote, err := t.target.Dial("tcp", fmt.Sprintf("%s:%d", t.RemoteHost, t.RemotePort))
	if err != nil {
		return err
	}
	defer func() {
		log.Printf("Closing remote connection...")
		remote.Close()
	}()

	// mirror the connections
	log.Printf("Forwarding localhost:%d -> %s:%d...", t.LocalPort, t.RemoteHost, t.RemotePort)
	go func() {
		_, err := io.Copy(local, remote)
		if err != nil {
			log.Printf("Error copying local to remote: %v", err)
		}
		done <- true
	}()
	go func() {
		_, err := io.Copy(remote, local)
		if err != nil {
			log.Printf("Error copying remote to local: %v", err)
		}
		done <- true
	}()

	// block while both connections are open
	<-done
	return nil
}
