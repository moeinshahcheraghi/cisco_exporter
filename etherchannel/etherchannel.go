package etherchannel

type EtherChannelGroup struct {
    Group       string
    PortChannel string
    Status      string
    Protocol    string
    Ports       []EtherChannelPort
}

type EtherChannelPort struct {
    Port   string
    Status string
}