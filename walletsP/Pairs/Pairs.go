package Pairs

type Pair struct {
	PairName      string
	Pair          uint16
	ZserverIP     string
	ZserverPort   uint16
	ZserverWSPort uint16
	ObookW        uint16
	ObookP        uint16
	ObookIP       string
	PairUrl       string
	Coin          uint16
	Market        uint16
}

type WalletType struct {
	WalletID uint64
	WalletName string
	WalletShort string
	litecoin string
	RpcPort uint16
	RpcUser string
	RpcPassw string
	GraphicIp string
	GraphicPort uint16

}

var WalletList [1000]WalletType
var PairList [1000]Pair
var Delimiter [1000]uint64

func init() {
	LoadPairs()
	LoadWAllets()
}

func LoadPairs() {
	PairList[4] = Pair{Coin: 4, Market: 1, PairName: "ETH/BTC", Pair: 4, ZserverIP: "127.0.0.1", ZserverPort: 5555, ZserverWSPort: 8001, ObookIP: "127.0.0.1", ObookW: 9985, ObookP: 9954, PairUrl: "ethbtc"}

	Delimiter[0] = 1000000000 //RVN
	Delimiter[2] = 1000000000  //EOS
	Delimiter[1] = 1000000000 //BTC
	Delimiter[4] = 1000000000 //ETH
	Delimiter[3] = 1000      //XRP
	Delimiter[0] = 1000      //RVN
	Delimiter[6] = 1000      //ETHO
	Delimiter[7] = 1000      //ETC
	Delimiter[8] = 1000      //DOGE
	Delimiter[9] = 1000      //LTC
}

func LoadWAllets() {
	WalletList[1] = WalletType{WalletID: 1, WalletName: "Bitcoin",  WalletShort: "BTC", litecoin: "127.0.0.1", RpcUser: "rpcclient",  RpcPassw: "P0", RpcPort: 19332, GraphicIp: "127.0.0.1", GraphicPort: 8104}
	WalletList[4] = WalletType{WalletID: 4, WalletName: "Ethereum",  WalletShort: "ETH", litecoin: "127.0.0.1", RpcUser: "rpcclient",  RpcPassw: "P0", RpcPort: 19332, GraphicIp: "127.0.0.1", GraphicPort: 8106}
}