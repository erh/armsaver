package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"go.viam.com/rdk/components/arm"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/referenceframe"

	"github.com/erh/vmodutils"
)

type data struct {
	All [][]referenceframe.Input
}

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}
func realMain() error {
	ctx := context.Background()
	logger := logging.NewLogger("armsaver")

	host := flag.String("host", "", "host")
	cmd := flag.String("cmd", "learn", "")
	debug := flag.Bool("debug", false, "")

	flag.Parse()

	if *debug {
		logger.SetLevel(logging.DEBUG)
	}

	if *host == "" {
		return fmt.Errorf("need host")
	}

	machine, err := vmodutils.ConnectToHostFromCLIToken(ctx, *host, logger)
	if err != nil {
		return err
	}
	defer machine.Close(ctx)

	myArm, err := arm.FromRobot(machine, "arm")
	if err != nil {
		return err
	}

	if *cmd == "learn" {
		inReader := bufio.NewReader(os.Stdin)

		logger.Infof("Press Enter to start (then enter to stop...")
		inReader.ReadBytes('\n')

		armContext, cancel := context.WithCancel(ctx)

		d := &data{}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			start := time.Now()
			defer wg.Done()
			for armContext.Err() == nil {
				in, err := myArm.JointPositions(ctx, nil)
				if err != nil {
					logger.Warnf("can't JointPositions: %v", err)
				} else {
					d.All = append(d.All, in)
					logger.Infof("%d - %v", len(d.All), time.Since(start))
				}

				time.Sleep(time.Millisecond * 10)
			}

		}()

		inReader.ReadBytes('\n')
		cancel()
		wg.Wait()

		jsonData, err := json.Marshal(d)
		if err != nil {
			return err
		}

		err = os.WriteFile("foo.json", jsonData, 0666)
		if err != nil {
			return err
		}
	} else if *cmd == "replay" {
		for _, fn := range flag.Args() {
			d := &data{}
			err = vmodutils.ReadJSONFromFile(fn, d)
			if err != nil {
				return err
			}

			err = myArm.MoveThroughJointPositions(ctx, d.All, &arm.MoveOptions{MaxVelRads: 20}, nil)
			if err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("unknown command: %v", *cmd)
	}

	return nil

}
