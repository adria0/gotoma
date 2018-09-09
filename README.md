[![](README.md.images/gotomabanner_small.png)]()

## Inspiration

Last two years I have been aware of **some problems that appeared when using the decentralized technologies**, and some of them are:

- **lack of privacy when interacting with services** (how infura manages my privacy? is etherscan tracking my web interactions). Can they know my accounts?

- **trusting in centralized services**... are public JSON-RPC providers answering the correct data? data visualization of my transactions in XXX is ok?

- **trusting only in the browser***... what about a malware that modifies the browser content, this really scares me a lot.

So, those services are really important but is also important to have other options, less centralized.

DAppNode (https://dappnode.io/, just got an EFC and Aragon grants) is a community Linux distro with a container-manager for decentralized applications, and it comes with the possibility to easy install a different kind of full nodes and develop applications that use these nodes, among other interesting options. So, developing an application for it seems to have a lot of sense.

## What it does

**The application is a server that continuously scans for dappnode-installed full nodes** (for the moment only ethereum nodes). **In this configuration, you can specify**:

- **the blockchains installed**,
- **accounts/smartcontracts** to monitor
-  **define flexible rules about when get notified**

When it occurs, **you will be notified via Telegram** (or another DAppNode installed messaging agent).

Here is an example of a configuration file:

```yaml
# configuration file
networks:
    ethmain:
        type: ethereum
        url: ws://my.ethchain.dnp.dappnode.eth:8546 

accounts:
    # simple account monitor, from or to
    0x137d9174d3bd00f2153dcc0fe7af712d3876a71e:
        network : ethmain

alerts:
    # scan smartcontract action
    createSiringAuction:
        network : ethmain
        rule: (to == '0x06012c8cf97bead5deae237070f9587f8e7a266d' && data =~ '0xf7d8c883')
        message: KryptoKittiy createSiringAuction called with gas {{ .gasprice }}
    
    # check events in all transactions if there’s my account in the 0x6e81… topic of the 0x7d13… contract
    deepanalisys:
        network : ethmain
        rule: (log_0x7d1335Af903ff256823c9AA2d4a5aaA41E054335_0x6e812926864597b1b871e35c4b24bd297ec1e96c871c41b9d7d3deb47bbe751c =~ '137d9174d3bd00f2153dcc0fe7af712d3876a71e')
            message: Somebody made me a transfer 

notifications:
    telegramusername: adriamb
```

Telegram sends a notification like:

`alert: Generic 0x137d9174d3bd00f2153dcc0fe7af712d3876a71e account modified http://my.gotoma.dnp.adriamb.eth:8080/b/ethmain/tx/0x8d3eb7670582df8e08ece5cda026fbb2bc4a1f5c7a93633744009249397ab399`

The `http://my.gotoma.dnp.adriamb.eth:8080/...` is a  `.eth` domain that can be accessed via DAppNode VPN integrated ENS resolution, there **you can see more information about the transaction done using the data from your full node, not from an external-centralized service**.
