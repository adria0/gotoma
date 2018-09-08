package config

import "io/ioutil"

// C is the package config
var C Config
var Valid bool

// Config is the server configurtion
type Config struct {
	Networks map[string]struct {
		Type string
		URL  string
	}

	Accounts map[string]struct {
		Network string
	}

	Alerts map[string]struct {
		Network string
		Rule    string
		Message string
	}

	Notifications struct {
		TelegramUsername string
	}
}

var Default = `
networks:
    ethmain:
        type: ethereum
        url: ws://my.ethchain.dnp.dappnode.eth:8546 
accounts:

    # simple account monitor, from or to
    # 0x137d9174d3bd00f2153dcc0fe7af712d3876a71e:
    #     network : ethmain

alerts:

    # createSiringAuction:
    #    network : ethmain
    #    rule: (to == '0x06012c8cf97bead5deae237070f9587f8e7a266d' && data =~ '0xf7d8c883')
    #    message: createSiringAuction called with gas {{ .gasprice }}

    # deepanalisys:
    #    network : ethmain
    #    rule: (log_0x7d1335Af903ff256823c9AA2d4a5aaA41E054335_0x6e812926864597b1b871e35c4b24bd297ec1e96c871c41b9d7d3deb47bbe751c =~ '0xf7d8c883')
    #    message: createSiringAuction called with gas {{ .gasprice }}

#notifications:
#	telegramusername: adriamb

`

func Get() string {
	content, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		return Default
	}
	return string(content)
}

func Set(content string) {
	ioutil.WriteFile("./config.yaml", []byte(content), 0744)
}
