package Library

import (
	"encoding/binary"
	//"exchange2/Pairs" //set ip adress and other settings
	zmq "github.com/pebbe/zmq4"
	"wallets/walletsP/Pairs"
)

var SendToObookCh = make(chan []byte, 1000000)
/*
function processBuySell
exchange
*/
func ProcessBuySell(isBuy bool, userId uint64, price uint64, amount uint64, pair int, session string, status uint8, currencies *[100][10000]uint64, delimiter [1000]uint64, globalCounter int, WriteTolog chan []byte, tcpPairs[1000]Pairs.Pair) int{
	return 0
	//if (userId > 10000){
	//	return 2
	//}
	//if (pair >= len(tcpPairs)){
	//	return 2
	//}
	//if (isBuy == true){
	//
	//	if(currencies[tcpPairs[pair].Market][userId] == 0){
	//		return 0
	//	}
	//	if(currencies[tcpPairs[pair].Market][userId] < uint64(amount*price/delimiter[tcpPairs[pair].Coin])){
	//		fmt.Println("currencies[pairs[pair].market][userId] < uint64(amount*price/delimiter[pairs[pair].coin])", currencies[tcpPairs[pair].Market][userId], uint64(amount*price/delimiter[tcpPairs[pair].Coin]))
	//		return 0
	//	}
	//	if status == 0 {
	//		//pairs[4] = Pair{0,1}
	//		currencies[tcpPairs[pair].Market][userId] = currencies[tcpPairs[pair].Market][userId] - uint64(amount*price/delimiter[tcpPairs[pair].Coin])
	//	}
	//	WriteTolog <- ConvertBalanceMsg(tcpPairs[pair].Market, userId, currencies[tcpPairs[pair].Market][userId])
	//
	//	if (globalCounter%100000 == 1){
	//		fmt.Println("buy", userId, price, amount, session, pair, len(tcpPairs), currencies[tcpPairs[pair].Market][userId])
	//	}
	//} else {
	//	if(currencies[tcpPairs[pair].Coin][userId] == 0){
	//		return 0
	//	}
	//	if(currencies[tcpPairs[pair].Coin][userId] < amount){
	//		return 0
	//	}
	//	if status == 0 {
	//		currencies[tcpPairs[pair].Coin][userId] = currencies[tcpPairs[pair].Coin][userId] - amount
	//	}
	//	WriteTolog <- ConvertBalanceMsg(tcpPairs[pair].Market, userId, currencies[tcpPairs[pair].Market][userId])
	//	if (globalCounter%100000 == 1){
	//		fmt.Println("sell", userId, price, amount, session, pair, len(tcpPairs), currencies[tcpPairs[pair].Coin][userId])
	//	}
	//}
	//return 1
}

/*
EXchange
*/
func SendDataToObook(SendToObook map[int]*zmq.Socket){
	//for{
	//	data, err  := <- SendToObookCh
	//	if err {
	//		fmt.Println(err)
	//	}
	//	pair := binary.LittleEndian.Uint16(data[4:6])
	//	pairInt := int(pair)
	//	SendToObook[pairInt].SendBytes(data,0)
	//}
}


/*
function convertToMsg
exchange
*/
func ConvertToMsg(uId uint32, pair uint16, status uint8, isBuy bool, oId uint64, oTime uint64, amount uint64, filledAmount uint64, price uint64) ([]byte){
	return make([]byte, 0)
	//lineBuf := make([]byte, 48)
	//
	//binary.LittleEndian.PutUint32(lineBuf[0:4], uId)
	//binary.LittleEndian.PutUint16(lineBuf[4:6], uint16(pair))
	//
	//var isBuyInt uint16
	//if (isBuy == true ){
	//	isBuyInt = 1
	//}else{
	//	isBuyInt = 0
	//}
	//binary.LittleEndian.PutUint16(lineBuf[6:8], (uint16(isBuyInt)<<15) | uint16(status))
	//binary.LittleEndian.PutUint64(lineBuf[8:16], uint64(oId))
	//binary.LittleEndian.PutUint64(lineBuf[16:24], uint64(oTime))
	//binary.LittleEndian.PutUint64(lineBuf[24:32], uint64(amount))
	//binary.LittleEndian.PutUint64(lineBuf[32:40], uint64(filledAmount))
	//binary.LittleEndian.PutUint64(lineBuf[40:48], uint64(price))
	//fmt.Println("Converted")
	//return lineBuf
}

/*
function convertBalanceMsg
exchange
*/
func ConvertBalanceMsg(currency uint16, userId uint64, balance uint64) []byte{
	lineBuf := make([]byte, 18)
	binary.LittleEndian.PutUint16(lineBuf[0:2], currency)
	binary.LittleEndian.PutUint64(lineBuf[2:10], userId)
	binary.LittleEndian.PutUint64(lineBuf[10:18], balance)

	return lineBuf
}
