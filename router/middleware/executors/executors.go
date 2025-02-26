// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package executors

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/go-vela/types/library"

	"github.com/go-vela/server/database"
	"github.com/go-vela/server/router/middleware/build"
	"github.com/go-vela/server/util"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Retrieve gets the executors in the given context.
func Retrieve(c *gin.Context) []library.Executor {
	return FromContext(c)
}

// Establish sets the executors in the given context.
func Establish() gin.HandlerFunc {
	return func(c *gin.Context) {
		e := new([]library.Executor)
		b := build.Retrieve(c)
		// retrieve the worker
		w, err := database.FromContext(c).GetWorker(b.GetHost())
		if err != nil {
			retErr := fmt.Errorf("unable to get worker: %w", err)
			util.HandleError(c, http.StatusNotFound, retErr)

			return
		}

		// prepare the request to the worker to retrieve executors
		client := http.DefaultClient
		client.Timeout = 30 * time.Second
		endpoint := fmt.Sprintf("%s/api/v1/executors", w.GetAddress())

		req, err := http.NewRequestWithContext(context.Background(), "GET", endpoint, nil)
		if err != nil {
			retErr := fmt.Errorf("unable to form request to %s: %w", endpoint, err)
			util.HandleError(c, http.StatusBadRequest, retErr)

			return
		}

		// add the token to authenticate to the worker as a header
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.MustGet("secret").(string)))

		// make the request to the worker and check the response
		resp, err := client.Do(req)
		if err != nil || resp == nil {
			// abandoned builds might have ran on a worker that no longer exists
			// if the worker is unavailable write an empty slice ToContext
			ToContext(c, *e)
			c.Next()
		}
		defer resp.Body.Close()

		// Read Response Body
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			retErr := fmt.Errorf("unable to read response from %s: %w", endpoint, err)
			util.HandleError(c, http.StatusBadRequest, retErr)

			return
		}

		// parse response and validate at least one item was returned
		err = json.Unmarshal(respBody, e)
		if err != nil {
			retErr := fmt.Errorf("unable to parse response from %s: %w", endpoint, err)
			util.HandleError(c, http.StatusBadRequest, retErr)

			return
		}

		ToContext(c, *e)
		c.Next()
	}
}
