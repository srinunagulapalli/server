// Copyright (c) 2021 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package api

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/go-vela/server/database"
	"github.com/go-vela/server/router/middleware/build"
	"github.com/go-vela/server/router/middleware/repo"
	"github.com/go-vela/server/router/middleware/service"
	"github.com/go-vela/server/router/middleware/step"
	"github.com/go-vela/server/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

// ServiceStream represents the API handler to do stuff...
func ServiceStream(c *gin.Context) {
	// capture middleware values
	b := build.Retrieve(c)
	r := repo.Retrieve(c)
	s := service.Retrieve(c)

	entry := fmt.Sprintf("%s/%d", r.GetFullName(), b.GetNumber())

	logrus.Infof("streaming logs for service %d for build %s", s.GetNumber(), entry)

	// upgrade the HTTP connection to WebSocket connection
	//
	// https://pkg.go.dev/github.com/gorilla/websocket#Upgrader.Upgrade
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		retErr := fmt.Errorf("unable to stream logs for service %s/%d: %w", entry, s.GetNumber(), err)

		util.HandleError(c, http.StatusInternalServerError, retErr)

		return
	}
	defer conn.Close()

	// create new buffer for uploading logs
	logs := new(bytes.Buffer)
	// create new channel for processing logs
	done := make(chan bool)
	// defer closing channel to stop processing logs
	defer close(done)

	// send API call to capture the service logs
	_log, err := database.FromContext(c).GetServiceLog(s.GetID())
	if err != nil {
		retErr := fmt.Errorf("unable to get logs for service %s/%d: %w", entry, s.GetNumber(), err)

		util.HandleError(c, http.StatusInternalServerError, retErr)

		return
	}

	go func() {
		logrus.Debugf("polling websocket buffer for service %d for build %s", s.GetNumber(), entry)

		// spawn "infinite" loop that will upload logs
		// from the buffer until the channel is closed
		for {
			// sleep for "1s" before attempting to upload logs
			time.Sleep(1 * time.Second)

			// create a non-blocking select to check if the channel is closed
			select {
			// after repo timeout of idle (no response) end the stream
			//
			// this is a safety mechanism
			case <-time.After(time.Duration(r.GetTimeout()) * time.Minute):
				logrus.Tracef("repo timeout of %d exceeded", r.GetTimeout())

				return
			// channel is closed
			case <-done:
				logrus.Trace("channel closed for polling container logs")

				// return out of the go routine
				return
				// channel is not closed
			default:
				// update the existing log with the new bytes
				//
				// https://pkg.go.dev/github.com/go-vela/types/library?tab=doc#Log.AppendData
				_log.AppendData(logs.Bytes())

				// send API call to update the log
				err = database.FromContext(c).UpdateLog(_log)
				if err != nil {
					retErr := fmt.Errorf("unable to update logs for service %s/%d: %w", entry, s.GetNumber(), err)

					util.HandleError(c, http.StatusInternalServerError, retErr)

					return
				}

				// flush the buffer of logs
				logs.Reset()
			}
		}
	}()

	logrus.Debugf("reading logs from websocket connection for service %d for build %s", s.GetNumber(), entry)

	for {
		// set timeout of 10s to send the logs
		//
		// https://pkg.go.dev/github.com/gorilla/websocket#Conn.SetReadDeadline
		err := conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			retErr := fmt.Errorf("unable to set timeout for websocket connection for step %s/%d: %w", entry, s.GetNumber(), err)

			util.HandleError(c, http.StatusInternalServerError, retErr)

			return
		}

		// read messages from the websocket connection
		//
		// https://pkg.go.dev/github.com/gorilla/websocket#Conn.ReadMessage
		_, message, err := conn.ReadMessage()
		if err != nil {
			// check if the websocket was closed
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				logrus.Tracef("websocket unexpectedly closed for service %d for build %s: %v", s.GetNumber(), entry, err)

				retErr := fmt.Errorf("unable to read logs from websocket connection for service %s/%d: %w", entry, s.GetNumber(), err)

				util.HandleError(c, http.StatusInternalServerError, retErr)

				return
			}

			logrus.Tracef("websocket closed for service %d for build %s: %v", s.GetNumber(), entry, err)

			break
		}

		// write all the logs from the websocket connection
		logs.Write(append(message, []byte("\n")...))
	}
}

// StepStream represents the API handler to do stuff...
func StepStream(c *gin.Context) {
	// capture middleware values
	b := build.Retrieve(c)
	r := repo.Retrieve(c)
	s := step.Retrieve(c)

	entry := fmt.Sprintf("%s/%d", r.GetFullName(), b.GetNumber())

	logrus.Infof("streaming logs for step %d for build %s", s.GetNumber(), entry)

	// upgrade the HTTP connection to WebSocket connection
	//
	// https://pkg.go.dev/github.com/gorilla/websocket#Upgrader.Upgrade
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		retErr := fmt.Errorf("unable to stream logs for step %s/%d: %w", entry, s.GetNumber(), err)

		util.HandleError(c, http.StatusInternalServerError, retErr)

		return
	}
	defer conn.Close()

	// create new buffer for uploading logs
	logs := new(bytes.Buffer)
	// create new channel for processing logs
	done := make(chan bool)
	// defer closing channel to stop processing logs
	defer close(done)

	// send API call to capture the step logs
	_log, err := database.FromContext(c).GetStepLog(s.GetID())
	if err != nil {
		retErr := fmt.Errorf("unable to get logs for step %s/%d: %w", entry, s.GetNumber(), err)

		util.HandleError(c, http.StatusInternalServerError, retErr)

		return
	}

	go func() {
		logrus.Debugf("polling websocket buffer for step %d for build %s", s.GetNumber(), entry)

		// spawn "infinite" loop that will upload logs
		// from the buffer until the channel is closed
		for {
			// sleep for "1s" before attempting to upload logs
			time.Sleep(1 * time.Second)

			// create a non-blocking select to check if the channel is closed
			select {
			// after repo timeout of idle (no response) end the stream
			//
			// this is a safety mechanism
			case <-time.After(time.Duration(r.GetTimeout()) * time.Minute):
				logrus.Tracef("repo timeout of %d exceeded", r.GetTimeout())

				return
			// channel is closed
			case <-done:
				logrus.Trace("channel closed for polling container logs")

				// return out of the go routine
				return
				// channel is not closed
			default:
				// update the existing log with the new bytes
				//
				// https://pkg.go.dev/github.com/go-vela/types/library?tab=doc#Log.AppendData
				_log.AppendData(logs.Bytes())

				// send API call to update the log
				err = database.FromContext(c).UpdateLog(_log)
				if err != nil {
					retErr := fmt.Errorf("unable to update logs for step %s/%d: %w", entry, s.GetNumber(), err)

					util.HandleError(c, http.StatusInternalServerError, retErr)

					return
				}

				// flush the buffer of logs
				logs.Reset()
			}
		}
	}()

	logrus.Debugf("reading logs from websocket connection for step %d for build %s", s.GetNumber(), entry)

	for {
		// set timeout of 10s to send the logs
		//
		// https://pkg.go.dev/github.com/gorilla/websocket#Conn.SetReadDeadline
		err := conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			retErr := fmt.Errorf("unable to set timeout for websocket connection for step %s/%d: %w", entry, s.GetNumber(), err)

			util.HandleError(c, http.StatusInternalServerError, retErr)

			return
		}

		// read messages from the websocket connection
		//
		// https://pkg.go.dev/github.com/gorilla/websocket#Conn.ReadMessage
		_, message, err := conn.ReadMessage()
		if err != nil {
			// check if the websocket was closed
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				logrus.Tracef("websocket unexpectedly closed for step %d for build %s: %v", s.GetNumber(), entry, err)

				retErr := fmt.Errorf("unable to read logs from websocket connection for step %s/%d: %w", entry, s.GetNumber(), err)

				util.HandleError(c, http.StatusInternalServerError, retErr)

				return
			}

			logrus.Tracef("websocket closed for step %d for build %s: %v", s.GetNumber(), entry, err)

			break
		}

		// write all the logs from the websocket connection
		logs.Write(append(message, []byte("\n")...))
	}
}
