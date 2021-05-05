package main

import (
	"cwlogsexample/datasource"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

var (
	cwl           *cloudwatchlogs.CloudWatchLogs
	logGroupName  = "youtubeTest"
	logStreamName = ""
	sequenceToken = ""
)

func init() {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("eu-west-2"), // london
		},
	})

	if err != nil {
		panic(err)
	}

	cwl = cloudwatchlogs.New(sess)

	err = ensureLogGroupExists(logGroupName)
	if err != nil {
		panic(err)
	}

}

func main() {
	queue := []string{}
	lock := sync.Mutex{}

	go datasource.GenerateData(&queue, &lock)

	go processQueue(&queue, &lock)

	// to stop the code from exiting
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()

}

// ensureLogGroupExists first checks if the log group exists,
// if it doesn't it will create one.
func ensureLogGroupExists(name string) error {
	resp, err := cwl.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{})
	if err != nil {
		return err
	}

	for _, logGroup := range resp.LogGroups {
		if *logGroup.LogGroupName == name {
			return nil
		}
	}

	_, err = cwl.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: &name,
	})
	if err != nil {
		return err
	}

	_, err = cwl.PutRetentionPolicy(&cloudwatchlogs.PutRetentionPolicyInput{
		RetentionInDays: aws.Int64(14),
		LogGroupName:    &name,
	})

	return err
}

// createLogStream will make a new logStream with a random uuid as its name.
func createLogStream() error {
	name := uuid.New().String()

	_, err := cwl.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  &logGroupName,
		LogStreamName: &name,
	})

	logStreamName = name

	return err
}

// processQueue will process the log queue
func processQueue(queue *[]string, lock *sync.Mutex) error {
	var logQueue []*cloudwatchlogs.InputLogEvent

	for {
		lock.Lock()
		if len(*queue) > 0 {
			for _, item := range *queue {
				logQueue = append(logQueue, &cloudwatchlogs.InputLogEvent{
					Message:   &item,
					Timestamp: aws.Int64(time.Now().UnixNano() / int64(time.Millisecond)),
				})
			}

			*queue = []string{}
		}

		lock.Unlock()

		if len(logQueue) > 0 {
			input := cloudwatchlogs.PutLogEventsInput{
				LogEvents:    logQueue,
				LogGroupName: &logGroupName,
			}

			if sequenceToken == "" {
				err := createLogStream()
				if err != nil {
					panic(err)
				}
			} else {
				input = *input.SetSequenceToken(sequenceToken)
			}

			input = *input.SetLogStreamName(logStreamName)

			resp, err := cwl.PutLogEvents(&input)
			if err != nil {
				log.Println(err)
			}

			if resp != nil {
				sequenceToken = *resp.NextSequenceToken
			}

			logQueue = []*cloudwatchlogs.InputLogEvent{}
		}

		time.Sleep(time.Second * 5)
	}

}
