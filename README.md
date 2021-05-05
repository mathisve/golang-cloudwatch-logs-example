# Golang CloudWatch Logs Example
![build](https://github.com/mathisve/golang-cloudwatch-logs-example/actions/workflows/go.yaml/badge.svg?branch=master)

Simple example explaining the basics of CloudWatch Logs.

## Sample 
```go
func createLogStream() error {
	name := uuid.New().String()

	_, err := cwl.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  &logGroupName,
		LogStreamName: &name,
	})

	logStreamName = name

	return err
}
```

## Step-By-Step Youtube Video
[![picture](https://github.com/mathisve/golang-cloudwatch-logs-example/blob/master/img/cloudwatch-logs.png)](https://youtu.be/aZ-gP4rbFDo)

## More Information
Find the Cloudwatch Logs Documentation [here](https://docs.aws.amazon.com/sdk-for-go/api/service/cloudwatchlogs/).