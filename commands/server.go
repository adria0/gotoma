package commands

import (
	"bytes"
	"fmt"
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
	kServerHost     = "my.tomahawk.dnp.dappnode.eth"
	kSendMessageUrl = ""
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

type logentry struct {
	netid   string
	alert   string
	message string
	tx      *types.Transaction
	r       *types.Receipt
}

func (l *logentry) MsgString() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("alert: %v ", l.alert))
	b.WriteString(fmt.Sprintf("%v ", l.message))
	b.WriteString(fmt.Sprintf("http://" + kServerHost + "/b/" + l.netid + "/tx/" + l.tx.Hash().Hex()))

	return b.String()
}

func (l *logentry) LogString() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("alert: %v\n", l.alert))
	b.WriteString(fmt.Sprintf("message: %v\n", l.message))
	b.WriteString(fmt.Sprintf("%v", l.tx.String()))

	return b.String()
}

var mutex sync.Mutex
var logtext string

func onLog(netid, alert, message string, tx *types.Transaction, r *types.Receipt) {
	mutex.Lock()
	defer mutex.Unlock()

	text := txText(tx, r)
	if len(logtext) > 8192 {
		logtext = text + "<hr>" + logtext[:8192-len(text)]
	} else {
		logtext = text + "<hr>" + logtext
	}

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

			c.String(http.StatusOK, networks[netid].TxInfo(txid))
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
