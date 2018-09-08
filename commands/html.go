package commands

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

var htmlmain = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>Tomahawk!</title>
<style type="text/css" media="screen">
    #editor { 
	   height: 300px
    }
</style>
</head>
<body>
<h1>Configuration</h1>
<div id="editor">{{ .Config  }}</div>

<script src="https://cdnjs.cloudflare.com/ajax/libs/ace/1.4.1/ace.js" integrity="sha256-kCykSp9wgrszaIBZpbagWbvnsHKXo4noDEi6ra6Y43w=" crossorigin="anonymous"></script>    
<script src="https://cdnjs.cloudflare.com/ajax/libs/ace/1.4.1/mode-yaml.js" integrity="sha256-95xNUgbfIXvRXJezV53+JM5HPO6PnJ+wZ7/GwdesKAE=" crossorigin="anonymous"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/2.2.4/jquery.min.js"></script>
<script>
var editor = ace.edit("editor");
editor.setTheme("ace/theme/monokai");
editor.session.setMode("ace/mode/yaml");

function save() {
    $.ajax({
        type: "POST",
        url: "/config",
        data: JSON.stringify({ config: editor.getValue()  }),
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function(data){
			alert("Saved. Must reload service.")
        },
        failure: function(errMsg) {
            alert(errMsg);
        }
    });
}

</script>
<button onclick="save()">Save config (needs reload service)</button>


<h1>Logs</h1>
{{ .Logs }}

</body>
</html>
`

func htmlTx(tx *types.Transaction, receipt *types.Receipt) string {
	var b bytes.Buffer

	cell := func(k string, v interface{}) {
		b.WriteString(fmt.Sprintf("<tr><td><b>%v</b></td><td>%v</td></tr>", k, v))
	}

	b.WriteString("<table>")
	if receipt.Status == types.ReceiptStatusSuccessful {
		cell("Status", "SUCCESS")
	} else {
		cell("Status", "FAILED")
	}

	signer := types.NewEIP155Signer(tx.ChainId())
	from, _ := signer.Sender(tx)

	cell("TX", tx.Hash().Hex())
	cell("From", from.Hex())
	if tx.To() == nil {
		cell("To", "CREATE_CONTRACT")
	} else {
		cell("To", tx.To().Hex())
	}
	cell("Value", tx.Value().String())
	cell("GasPrice", tx.GasPrice().String())
	cell("GasLimit", tx.GasPrice().String())
	cell("Nonce", tx.Nonce())
	cell("Data", hex.EncodeToString(tx.Data()))
	cell("GasUsed", receipt.GasUsed)
	cell("CumulativeGasUsed", receipt.CumulativeGasUsed)
	b.WriteString("</table>")

	return b.String()
}
