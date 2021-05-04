package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/google/uuid"
	"log"
	"time"

	"cwlogsexample/datasource"
	"sync"
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
			Region: aws.String("eu-west-2"),
		},
	})

	if err != nil {
		log.Panic(err)
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

	// keeps the code from exiting
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

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
		LogGroupName: &logGroupName,
	})

	return err
}

func createLogStream() error {
	name := uuid.New().String()

	_, err := cwl.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  &logGroupName,
		LogStreamName: &name,
	})

	logStreamName = name

	return err
}

func processQueue(queue *[]string, lock *sync.Mutex) {
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
