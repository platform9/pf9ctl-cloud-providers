package pmk

import (
	"fmt"

	"github.com/platform9/pf9ctl/pkg/color"
	"github.com/platform9/pf9ctl/pkg/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	iamAws "github.com/aws/aws-sdk-go/service/iam"
)

func CheckAmazonPovider(awsIamUser, awsID, awsSecret, awsRegion string) error {

	//creates a new session using static credentials that the user passes and "" stands for token that is of no use in this situation

	svc := iamAws.New(session.New(
		&aws.Config{
			Region:      aws.String(awsRegion),
			Credentials: credentials.NewStaticCredentials(awsID, awsSecret, ""),
		}))

	//creates user input using username that is passed
	inputUser := &iamAws.GetUserInput{
		UserName: aws.String(awsIamUser),
	}

	//gets the user with the passed username
	resultUser, errUser := svc.GetUser(inputUser)
	if errUser != nil {
		return fmt.Errorf("Amazon user error " + errUser.Error())
	}
	arn := resultUser.User.Arn

	//checkpermission function takes the arn, svc and an array of permissions needed and checks if the user has all of them
	//this can be easily upgraded by just copy pasting one checkPermission block of code and changing the permission array
	if checkPermissions(arn, svc, util.EBSPermissions) {
		fmt.Println(color.Green("✓ ") + "ELB Access")
	} else {
		return fmt.Errorf(color.Red("X ") + "ELB Access Error ")
	}

	if checkPermissions(arn, svc, util.Route53Permissions) {
		fmt.Println(color.Green("✓ ") + "Route53 Access")
	} else {
		return fmt.Errorf(color.Red("X ") + "Route53 Access Error")
	}

	zoneSess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsID, awsSecret, ""),
	})

	if err != nil {
		return fmt.Errorf(color.Red("X ") + "Amazon client error " + err.Error())
	}

	zoneSvc := ec2.New(zoneSess)

	resultAvalZones, err := zoneSvc.DescribeAvailabilityZones(nil)
	if err != nil {
		return fmt.Errorf(color.Red("X ") + "Availability Zones error" + err.Error())
	}

	//checks if the user less than 2 availability zones
	if len(resultAvalZones.AvailabilityZones) < 2 {
		return fmt.Errorf("Availability Zones error, Minimum 2 availabiliy zones required but found %d", len(resultAvalZones.AvailabilityZones))

	}
	fmt.Println(color.Green("✓ ") + "Availability Zones success")

	if checkPermissions(arn, svc, util.EC2Permission) {
		fmt.Println(color.Green("✓ ") + "EC2 Access")
	} else {
		return fmt.Errorf(color.Red("X ") + "EC2 Access Error")

	}

	if checkPermissions(arn, svc, util.VPCPermission) {
		fmt.Println(color.Green("✓ ") + "VPC Access")
	} else {
		return fmt.Errorf(color.Red("X ") + "VPC Access Error")
	}

	if checkPermissions(arn, svc, util.IAMPermissions) {
		fmt.Println(color.Green("✓ ") + "IAM Access")
	} else {
		return fmt.Errorf(color.Red("X ") + "IAM Access Error")
	}

	if checkPermissions(arn, svc, util.AutoScalingPermissions) {
		fmt.Println(color.Green("✓ ") + "Autoscaling Access")
	} else {
		return fmt.Errorf(color.Red("X ") + "Autoscaling Access Error")

	}

	if checkPermissions(arn, svc, util.EKSPermissions) {
		fmt.Println(color.Green("✓ ") + "EKS Access")
	} else {
		return fmt.Errorf(color.Red("X ") + "EKS Access Error")
	}

	return nil

}

func checkPermissions(arn *string, svc *iamAws.IAM, actions []string) bool {

	//turns the array of strings into an array of pointers
	//this is done so it is easier to call checkpermissions since permissions can be pasted as strings
	//otherwise there would need to be more code to call this function
	actionNames := getActionNames(actions)

	//simulate gets all permissions from the actionNames array
	input := &iamAws.SimulatePrincipalPolicyInput{
		PolicySourceArn: arn,
		ActionNames:     actionNames,
	}

	result, err := svc.SimulatePrincipalPolicy(input)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	//and then checks if the user is allowed to use them by calling checkArray
	if !checkIfAllowed(result.EvaluationResults) {
		fmt.Println("Access Error")
		return false
	}

	return true

}

func checkIfAllowed(results []*iamAws.EvaluationResult) bool {

	//takes an array of user permissions and checks if the EvalDecision flag is not equal to allowed in which case the user doenst have
	//the permisison
	for i := range results {
		if *results[i].EvalDecision != "allowed" {
			return false
		}
	}
	return true

}

func getActionNames(actions []string) []*string {

	//turns array of strings into array of pointers
	var actionNames []*string
	for i := range actions {
		actionNames = append(actionNames, &actions[i])
	}
	return actionNames

}