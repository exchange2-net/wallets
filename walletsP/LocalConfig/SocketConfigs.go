package LocalConfig

//Ethereum
const ETHBalancePort = 9994
const ETHrcvNewAddressPort = 7104
const ETHTrnsctnHistoryPort = 7994


//Bitcoin server ports
const BTCBalancePort = 9992
const BTCOutTrnsctnPort = 7661
const BTCrcvNewAddressPort = 7101
const BTCsndTrnsctnHstrPort = 7991
const BTCsndTrnsctnHstrOUTPORT = 7891
const BTCsndColdTxInOUTPort = 7892

//Bitcoin OUT server ports
const BTCBalanceOUTPort = 9999
const BTCrcvTrnsctnHstrOUTPort = 7891
const BTCrcvColdTxInPort = 7892
const BTCsndOutTrnsctnPort = 7661
const BTCsndTrnsctnAnswerPort = 7411
const BTCrcvNewAddressOUTPort = 7201

//Wallet server Ports
const WalletaddressMapBegin = 7101
const WalletaddressMapEnd = 7110
const WalletaddressTrMapBegin = 7201
const WalletaddressTrMapEnd = 7210
const WalletrcvHistoryBegin = 7991
const WalletrcvHistoryEnd = 7999
const WalletrcvCheckBalancePort = 7302
const WalletsndCheckBalancePort = 7301
const WalletsendMailPort = 3702

//Mail server Ports
const MailhandleAUTHPort = 3701
const MailhandleWallets = 3702

//Auth server ports
const AuthSendToBalancePort = 3554
const AuthSendToWalletPort = 3650
const AuthSendToUContPort = 3651
const AuthmailPort = 3701

//User Contacts server ports
const UContactsAuthPort = 3651
const UcontactsHandlePort = 21001
const UcontactsUploadPort = ":21002"

//Balance server ports
const BalancercvCheckBlncPort = 7301
const BalancesndCheckBlncPort = 7302
const BalancesndBalance = "tcp://*:9995"
const BalancercvBalanceBegin = 9991
const BalancercvBalanceend = 9999
const BalancercvAccountAuthPort = 3554