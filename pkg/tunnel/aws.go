package tunnel

import (
        "math/rand"
        "time"
        "fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2instanceconnect"
)

type Instance struct {
	ID               string
	AvailabilityZone string
}

func getJumpInstance() (*Instance, error) {
	s := session.Must(session.NewSession())
	svc := ec2.New(s)

        // autodiscover running instances with the TuneJumpHost tag
        // XXX: ignores pagination for now - we don't expect many instances
	out, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag-key"),
				Values: aws.StringSlice([]string{"TuneJumpHost"}),
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: aws.StringSlice([]string{"running"}),
			},
		},
	})
	if err != nil {
		return nil, err
	}
        if len(out.Reservations) == 0 {
		return nil, fmt.Errorf("Could not find any active jump hosts")
        }

        // pick a random instance
        // XXX: ignores multiple instances per reservation - we only expect one
        rand.Seed(time.Now().UnixNano())
        randIdx := rand.Intn(len(out.Reservations))
        instance := out.Reservations[randIdx].Instances[0]

	return &Instance{
		ID:               *instance.InstanceId,
		AvailabilityZone: *instance.Placement.AvailabilityZone,
	}, nil
}

func sendKey(publicKey []byte, instanceId string, availabilityZone string) error {
	s := session.Must(session.NewSession())
	svc := ec2instanceconnect.New(s)

        // send the public key to the EC2 instance, where it's valid for 60s
	out, err := svc.SendSSHPublicKey(&ec2instanceconnect.SendSSHPublicKeyInput{
		AvailabilityZone: aws.String(availabilityZone),
		InstanceId:       aws.String(instanceId),
		InstanceOSUser:   aws.String("ec2-user"),
		SSHPublicKey:     aws.String(string(publicKey)),
	})
	if err != nil {
		return err
        }
        if out.Success == nil || !*out.Success {
		return fmt.Errorf("Failed to send public key to %s", instanceId)
	}

        return nil
}
