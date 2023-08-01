package tasks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/hibiken/asynq"
)

func Send_SMS_Worker(c context.Context, t *asynq.Task) error {

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://communication.test.ideopay.in/api/sendCommunication", bytes.NewReader(t.Payload()))
	if err != nil {
		return err
	}

	req.Header.Set("x-request-id", "56329838788")
	req.Header.Set("Content-Type", "application/json")
	fmt.Println("making http call")

	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("REQUEST:\n%s", string(reqDump))
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("error sending SMS: %s , %s", res.Status)

	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	t.ResultWriter().Write(b)
	fmt.Println("No Error")

	return nil
}
