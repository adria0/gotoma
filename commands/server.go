package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"text/template"

	cfg "github.com/TomahawkEthBerlin/gotoma/config"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

const (
	kEthereum       = "ethereum"
	kServerHost     = "my.gotoma.dnp.adriamb.eth:8080"
	kSendMessageUrl = "my.telegrambot.dnp.dappnode.eth"
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

var mutex sync.Mutex
var logtext string

func sendNotification(msg string) error {

	url := "http://" + kSendMessageUrl

	params := make(map[string]string)
	params["username"] = cfg.C.Notifications.TelegramUsername
	params["message"] = msg
	paramBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(paramBytes))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	return err
}

func onLog(netid, alert, message string, tx *types.Transaction, r *types.Receipt) {
	mutex.Lock()
	defer mutex.Unlock()

	text := htmlTx(tx, r)
	if len(logtext) > 8192 {
		logtext = text + "<hr>" + logtext[:8192-len(text)]
	} else {
		logtext = text + "<hr>" + logtext
	}

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("alert: %v ", alert))
	msg.WriteString(fmt.Sprintf("%v ", message))
	msg.WriteString(fmt.Sprintf("http://" + kServerHost + "/b/" + netid + "/tx/" + tx.Hash().Hex()))

	go func() {
		err := sendNotification(msg.String())
		if err != nil {
			log.Error(err)
		}
	}()
}

func Serve(cmd *cobra.Command, args []string) {

	r := gin.Default()

	if cfg.Valid {

		networks := make(map[string]Network)

		storage, err := NewKVStorage("./state")
		assert(err)

		for networkid, netinfo := range cfg.C.Networks {
			if netinfo.Type != kEthereum {
				assert(fmt.Errorf("Unknown network type '%v'", netinfo.Type))
			}
			networks[networkid] = NewEthNetwork(networkid, netinfo.URL, storage, onLog)
		}

		for networkid, network := range networks {
			log.Info("Starting dispatcher for ", networkid)
			network.Start()
		}

		r.GET("/b/:netid/tx/:txid", func(c *gin.Context) {
			netid := c.Param("netid")
			txid := c.Param("txid")

			c.Data(http.StatusOK, "text/html", []byte(networks[netid].TxInfo(txid)))
		})
	}

	r.POST("/config", func(c *gin.Context) {
		var json struct {
			Config string `json:"config" binding:"required"`
		}

		if c.BindJSON(&json) == nil {
			cfg.Set(json.Config)
		}
	})

	r.GET("/", func(c *gin.Context) {
		tmpl, _ := template.New("").Parse(htmlmain)
		var html bytes.Buffer
		tmpl.Execute(&html, struct {
			Config string
			Logs   string
		}{cfg.Get(), logtext})

		c.Data(http.StatusOK, "text/html; charset=utf-8", html.Bytes())
	})
	r.Run(":8080") // listen and serve on 0.0.0.0:8080
}
