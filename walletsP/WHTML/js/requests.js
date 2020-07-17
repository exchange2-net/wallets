var $ = jQuery.noConflict();
var HOST_IP = "localhost";
var HOST_PORT = ":80";
var session = "";
var userId = "";
var cookie_login = "";
var UserName = "";
var CoinID = [];
var Coin2ID = [];
var B_ID = [];
var CoinById = [];
var WalletsTemplate = "";
var TepmpalteDIV = "";
var ourFee = 0;
var CoinsPriceUSD = [];
var chart_bg_color = "#ffffff";
var chart_text_color = "#354052";
var feed = [];
var run = [];
var widgetOptions = {};
var subscription;
var lastBar = {
    close: 0
};
var ImgUrlToUpload = "";
var CurrentPage = window.location.pathname;
var CurrentPage = CurrentPage.replace("/", "");

if (CurrentPage == "") {
    const config = {
        supported_resolutions: ["1", "3", "5", "15", "30", "60", "120", "240", "D"]
    };

    var Datafeed = {
        onReady: cb => {
            setTimeout(() => cb(config), 0)

        },
        searchSymbols: (userInput, exchange, symbolType, onResultReadyCallback) => {},
        resolveSymbol: (symbolName, onSymbolResolvedCallback, onResolveErrorCallback) => {
            var symbol_stub = {
                name: symbolName,
                description: '',
                type: 'crypto',
                session: '24x7',
                timezone: 'Europe/Athens',
                ticker: symbolName,
                //exchange: split_data[0],
                minmov: 1,
                pricescale: 1,
                has_intraday: true,
                supported_resolution: config.supportedResolutions,
                volume_precision: 1,
                data_status: 'streaming',
            };

            setTimeout(function () {
                onSymbolResolvedCallback(symbol_stub)
            }, 0)

        },
        getBars: function (symbolInfo, resolution, from, to, onHistoryCallback, onErrorCallback, firstDataRequest) {
            var res;
            if (resolution == NaN) {
                res = "";
            } else {
                res = resolution;
            }
            feed = [];
            run.push(function () {
                onHistoryCallback(feed, {
                    noData: false
                });
            });
        },
        subscribeBars: (symbolInfo, resolution, onRealtimeCallback, subscribeUID, onResetCacheNeededCallback) => {
            subscription = onRealtimeCallback
        },
        unsubscribeBars: subscriberUID => {},
        calculateHistoryDepth: (resolution, resolutionBack, intervalBack) => {
            return resolution < 60 ? {
                resolutionBack: 'D',
                intervalBack: '10'
            } : undefined
        },
        getMarks: (symbolInfo, startDate, endDate, onDataCallback, resolution) => {
            //optional
        },
        getTimeScaleMarks: (symbolInfo, startDate, endDate, onDataCallback, resolution) => {
            //optional
        },
        getServerTime: cb => {}
    };

    widgetOptions = {
        debug: true,
        symbol: "BTC" + '/' + "USD",
        datafeed: Datafeed, // our datafeed object
        interval: '1',
        container_id: 'chart',
        library_path: '/charting_library/charting_library/',
        locale: 'en',
        timezone: 'Europe/Athens',
        disabled_features: ['use_localstorage_for_settings', 'volume_force_overlay', 'move_logo_to_main_pane', 'timeframes_toolbar'],
        enabled_features: ['support_multicharts', 'chart_property_page_scales'],
        client_id: 'test',
        user_id: 'public_user_id',
        fullscreen: false,
        autosize: true,
        overrides: {
            "paneProperties.background": chart_bg_color, //dark background
            "paneProperties.vertGridProperties.color": "#d8d8d8",
            "paneProperties.horzGridProperties.color": "#d8d8d8",
            "symbolWatermarkProperties.transparency": 0,
            "scalesProperties.textColor": chart_text_color,
            "mainSeriesProperties.candleStyle.wickUpColor": '#39b54a',
            "mainSeriesProperties.candleStyle.wickDownColor": '#7f323f',
            "volumePaneSize": "small",
        }
    };

    //TradingView.onready(function () {
    //   var widget = window.tvWidget = new TradingView.widget(widgetOptions);
    //});
}

function SHA256(s) {
    var chrsz = 8;
    var hexcase = 0;

    function safe_add(x, y) {
        var lsw = (x & 0xFFFF) + (y & 0xFFFF);
        var msw = (x >> 16) + (y >> 16) + (lsw >> 16);
        return (msw << 16) | (lsw & 0xFFFF);
    }

    function S(X, n) {
        return (X >>> n) | (X << (32 - n));
    }

    function R(X, n) {
        return (X >>> n);
    }

    function Ch(x, y, z) {
        return ((x & y) ^ ((~x) & z));
    }

    function Maj(x, y, z) {
        return ((x & y) ^ (x & z) ^ (y & z));
    }

    function Sigma0256(x) {
        return (S(x, 2) ^ S(x, 13) ^ S(x, 22));
    }

    function Sigma1256(x) {
        return (S(x, 6) ^ S(x, 11) ^ S(x, 25));
    }

    function Gamma0256(x) {
        return (S(x, 7) ^ S(x, 18) ^ R(x, 3));
    }

    function Gamma1256(x) {
        return (S(x, 17) ^ S(x, 19) ^ R(x, 10));
    }

    function core_sha256(m, l) {
        var K = new Array(0x428A2F98, 0x71374491, 0xB5C0FBCF, 0xE9B5DBA5, 0x3956C25B, 0x59F111F1, 0x923F82A4, 0xAB1C5ED5, 0xD807AA98, 0x12835B01, 0x243185BE, 0x550C7DC3, 0x72BE5D74, 0x80DEB1FE, 0x9BDC06A7, 0xC19BF174, 0xE49B69C1, 0xEFBE4786, 0xFC19DC6, 0x240CA1CC, 0x2DE92C6F, 0x4A7484AA, 0x5CB0A9DC, 0x76F988DA, 0x983E5152, 0xA831C66D, 0xB00327C8, 0xBF597FC7, 0xC6E00BF3, 0xD5A79147, 0x6CA6351, 0x14292967, 0x27B70A85, 0x2E1B2138, 0x4D2C6DFC, 0x53380D13, 0x650A7354, 0x766A0ABB, 0x81C2C92E, 0x92722C85, 0xA2BFE8A1, 0xA81A664B, 0xC24B8B70, 0xC76C51A3, 0xD192E819, 0xD6990624, 0xF40E3585, 0x106AA070, 0x19A4C116, 0x1E376C08, 0x2748774C, 0x34B0BCB5, 0x391C0CB3, 0x4ED8AA4A, 0x5B9CCA4F, 0x682E6FF3, 0x748F82EE, 0x78A5636F, 0x84C87814, 0x8CC70208, 0x90BEFFFA, 0xA4506CEB, 0xBEF9A3F7, 0xC67178F2);
        var HASH = new Array(0x6A09E667, 0xBB67AE85, 0x3C6EF372, 0xA54FF53A, 0x510E527F, 0x9B05688C, 0x1F83D9AB, 0x5BE0CD19);
        var W = new Array(64);
        var a, b, c, d, e, f, g, h, i, j;
        var T1, T2;
        m[l >> 5] |= 0x80 << (24 - l % 32);
        m[((l + 64 >> 9) << 4) + 15] = l;
        for (var i = 0; i < m.length; i += 16) {
            a = HASH[0];
            b = HASH[1];
            c = HASH[2];
            d = HASH[3];
            e = HASH[4];
            f = HASH[5];
            g = HASH[6];
            h = HASH[7];
            for (var j = 0; j < 64; j++) {
                if (j < 16) W[j] = m[j + i];
                else W[j] = safe_add(safe_add(safe_add(Gamma1256(W[j - 2]), W[j - 7]), Gamma0256(W[j - 15])), W[j - 16]);
                T1 = safe_add(safe_add(safe_add(safe_add(h, Sigma1256(e)), Ch(e, f, g)), K[j]), W[j]);
                T2 = safe_add(Sigma0256(a), Maj(a, b, c));
                h = g;
                g = f;
                f = e;
                e = safe_add(d, T1);
                d = c;
                c = b;
                b = a;
                a = safe_add(T1, T2);
            }
            HASH[0] = safe_add(a, HASH[0]);
            HASH[1] = safe_add(b, HASH[1]);
            HASH[2] = safe_add(c, HASH[2]);
            HASH[3] = safe_add(d, HASH[3]);
            HASH[4] = safe_add(e, HASH[4]);
            HASH[5] = safe_add(f, HASH[5]);
            HASH[6] = safe_add(g, HASH[6]);
            HASH[7] = safe_add(h, HASH[7]);
        }
        return HASH;
    }

    function str2binb(str) {
        var bin = Array();
        var mask = (1 << chrsz) - 1;
        for (var i = 0; i < str.length * chrsz; i += chrsz) {
            bin[i >> 5] |= (str.charCodeAt(i / chrsz) & mask) << (24 - i % 32);
        }
        return bin;
    }

    function Utf8Encode(string) {
        string = string.replace(/\r\n/g, "\n");
        var utftext = "";
        for (var n = 0; n < string.length; n++) {
            var c = string.charCodeAt(n);
            if (c < 128) {
                utftext += String.fromCharCode(c);
            } else if ((c > 127) && (c < 2048)) {
                utftext += String.fromCharCode((c >> 6) | 192);
                utftext += String.fromCharCode((c & 63) | 128);
            } else {
                utftext += String.fromCharCode((c >> 12) | 224);
                utftext += String.fromCharCode(((c >> 6) & 63) | 128);
                utftext += String.fromCharCode((c & 63) | 128);
            }
        }
        return utftext;
    }

    function binb2hex(binarray) {
        var hex_tab = hexcase ? "0123456789ABCDEF" : "0123456789abcdef";
        var str = "";
        for (var i = 0; i < binarray.length * 4; i++) {
            str += hex_tab.charAt((binarray[i >> 2] >> ((3 - i % 4) * 8 + 4)) & 0xF) +
                hex_tab.charAt((binarray[i >> 2] >> ((3 - i % 4) * 8)) & 0xF);
        }
        return str;
    }
    s = Utf8Encode(s);
    return binb2hex(core_sha256(str2binb(s), s.length * chrsz));
}

function validateEmail(email) {
    var re = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
    return re.test(String(email).toLowerCase());
}

function setCookie(name, value, days) {
    var expires = "";
    if (days) {
        var date = new Date();
        date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
        expires = "; expires=" + date.toUTCString();
    }
    document.cookie = name + "=" + (value || "") + expires + "; path=/";
}

function getCookie(name) {
    var nameEQ = name + "=";
    var ca = document.cookie.split(';');
    for (var i = 0; i < ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0) == ' ') c = c.substring(1, c.length);
        if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length, c.length);
    }
    return null;
}

function eraseCookie(name) {
    document.cookie = name + '=; Max-Age=-99999999;';
}

function check_cookie() {
    var check_cookie_session = null;
    var check_cookie_userId = null;
    var check_cookie_userLogin = null;

    check_cookie_session = getCookie("wallets_u_s");
    check_cookie_userId = getCookie("wallets_u_id");
    check_cookie_userLogin = getCookie("wallets_u_log");

    if (check_cookie_userId != null && check_cookie_session != null && check_cookie_userLogin != null) {
        session = check_cookie_session;
        userId = check_cookie_userId;
        cookie_login = check_cookie_userLogin;
        LogIN();
    } else {
        if (window.location.pathname != "/" || window.location.pathname != "/sign-up.html") {
            if (window.location.pathname == "/") {
                return;
            } else if (window.location.pathname == "/sign-up.html") {
                return;
            }
            window.location.replace("http://" + HOST_IP + "");
        }
    }
}

function registration() {
    var login = $("#sign-up-email").val();
    var password = $("#sign-up-password").val();
    $("#modal__desc").empty();

    $("#sign-up-email").css("border", "1px solid #000");
    $("#sign-up-password").css("border", "1px solid #000");

    if (login && password) {
        if (login.length > 5 && password.length > 5) {
            var flag1 = validateEmail(login);
            if (flag1 == false) {
                $("#attention3").addClass("active");
                $("#modal_overlay").addClass("active");
                $("#modal__desc").append("Registration error! Wrong E-mail.");
                return;
            }
        } else {
            if (login.length < 5) {
                $("#sign-up-email").css("border", "1px solid rgb(193, 30, 15)");
            }
            if (password.length < 5) {
                $("#sign-up-password").css("border", "1px solid rgb(193, 30, 15)");
            }

            $("#attention3").addClass("active");
            $("#modal_overlay").addClass("active");
            $("#modal__desc").append("Registration error! Login or password is to shoort!");
            return;
        }
    } else {
        $("#sign-up-email").css("border", "1px solid rgb(193, 30, 15)");
        $("#sign-up-password").css("border", "1px solid rgb(193, 30, 15)");
        $("#attention3").addClass("active");
        $("#modal_overlay").addClass("active");
        $("#modal__desc").append("Registration error!");
        return;
    }

    var formData = [{
        name: 'login',
        value: login
    }, {
        name: 'password',
        value: password
    }];

    $.post("http://" + HOST_IP + ":8008/reg", formData, function (data) {
        if (data == '-1') {
            $("#sign-up-email").css("border", "1px solid rgb(193, 30, 15)");
            $("#sign-up-password").css("border", "1px solid rgb(193, 30, 15)");
            $("#attention3").addClass("active");
            $("#modal_overlay").addClass("active");
            $("#modal__desc").append("Registration error!");
        } else {
            [session, userId] = data.split(",");
            setCookie("wallets_u_s", session, 0.2); //session for 1 day
            setCookie("wallets_u_id", userId, 0.2); //session for 1 day
            setCookie("wallets_u_log", login, 0.2); //
            cookie_login = login;
            window.location.replace("http://" + window.location.host + "/dashboard.html")
        }
    });
}

function removeClass() {
    document.querySelector(".modal.active").classList.remove("active");
    document.querySelector(".modal-overlay.active").classList.remove("active");
}

function LogIN() {
    if (session.length != 0 && session != null && userId.length != 0 && userId != null && cookie_login != null) {
        if (window.location.pathname == "" || window.location.pathname == "/sign-up.html") {
            if (window.location.pathname == "") {
                window.location.replace("http://" + HOST_IP + "/dashboard.html");
            } else if (window.location.pathname == "/sign-up.html") {
                window.location.replace("http://" + HOST_IP + "/dashboard.html");
            }
        }
        UserName = getCookie("wallets_u_log")
    } else {
        login = $("#log-in-email").val();
        password = $("#log-in-password").val();

        $("#log-in-email").css("border", "1px solid #000");
        $("#log-in-password").css("border", "1px solid #000");

        var flag1 = validateEmail(login);
        if (flag1 == false) {
            $("#log-in-email").css("border", "1px solid rgb(193, 30, 15)");
            $("#attention3").addClass("active");
            $("#modal_overlay").addClass("active");
            $("#modal__desc").append("Login error!");
            return;
        }
        if (login.length < 5) {
            $("#log-in-email").css("border", "1px solid rgb(193, 30, 15)");
            $("#attention3").addClass("active");
            $("#modal_overlay").addClass("active");
            $("#modal__desc").append("Login is too short!");
            return;
        }
        if (password.length < 5) {
            $("#log-in-password").css("border", "1px solid rgb(193, 30, 15)");
            $("#attention3").addClass("active");
            $("#modal_overlay").addClass("active");
            $("#modal__desc").append("Password is too short!");
            return;
        }

        var formData = [{
            name: 'auth',
            value: SHA256(login + ' ' + password)
        }, {
            name: 'email',
            value: login
        }, {
            name: 'AuthType',
            value: 'wallet'
        }];
        $("#modal__desc").empty();
        jQuery.post("http://" + HOST_IP + ":8008/auth", formData, function (data) {
            googleAuth = data.split(" ");
            if (googleAuth[0].length == 10 && googleAuth[0] == "googleAuth") {
                window.location.replace(googleAuth[1]);
                return;
            }
            if (data == '-2') {
                $("#attention3").addClass("active");
                $("#modal_overlay").addClass("active");
                $("#modal__desc").append("Account was banned");
                return;
            } else if (data == '-1') {
                $("#attention3").addClass("active");
                $("#modal_overlay").addClass("active");
                $("#modal__desc").append("Login error! Try Again or contact support");

                return;
            } else {
                var results;
                results = data.split(",");

                session = results[0];
                userId = results[1];
                BTCWallet = results[2];

                var cookie_session = session;
                var cookie_userId = userId;

                var check_cookie_session = null;
                var check_cookie_userId = null;

                check_cookie_session = getCookie("wallets_u_s");
                check_cookie_userId = getCookie("wallets_u_id");

                setCookie("wallets_u_s", cookie_session, 0.2); //session for 1 day
                setCookie("wallets_u_id", cookie_userId, 0.2); //session for 1 day
                setCookie("wallets_u_log", login, 0.2); //session for 1 day
            }
            window.location.replace("http://" + HOST_IP + "/dashboard.html");
        });
    }
}

function logout() {
    var formData = [];
    formData.push({
        name: 'userId',
        value: userId
    });
    formData.push({
        name: 'session',
        value: session
    });
    $.post("http://" + HOST_IP + ":8008/logout", formData, function (data) {
        console.log(data)
        eraseCookie("wallets_u_s");
        eraseCookie("wallets_u_id");
        eraseCookie("wallets_u_log");
        session = "";
        userId = "";
        window.location.replace("http://" + HOST_IP + "");
    });

}

function replaceOnPage() {
    document.getElementById("UserName").innerText = UserName;
}

function Delete_wallet(coinID, walletAddress, ID) {
    if (document.getElementById(ID).querySelectorAll("input")[1]) {
        var formData = {
            data: "DeleteWallet" + "," + session + "," + userId + "," + "," + "," + ",",
            CoinID: coinID,
            walletAddress: walletAddress,
        };
        var divID = "wallet_" + coinID;
        jQuery.post("http://" + HOST_IP + ":20001", formData, function (data) {
            if (data == "1") {
                // document.getElementById(divID).innerHTML = "";
                document.getElementById(divID).remove();
                $("#add_wallet_button").css("display", "flex");

            } else {
                alert("Error");
            }
        });
    }
}

function create_wallet() {
    var coinIs = $('#create-wallet-select').find(":selected").val();
    var createdAddress;
    var formData = {
        data: "createWallet" + "," + session + "," + userId + "," + "," + "," + ",",
        coinType: coinIs,
        CoinID: CoinID[coinIs],
    };

    jQuery.post("http://" + HOST_IP + ":20001", formData, function (data) {
        if (data == "") {
            createdAddress = "Error";
            $("#created_wallet_address").val(createdAddress);
            return;
        }
        [createdAddress, coinid] = data.split(",");
        $("#created_wallet_address").val(createdAddress);
        $("#create-wallet-button").text("Created");
        $("#create-wallet-button").css('background-color', '#05c7e5');
        setTimeout(function () {
            $("#create-wallet-button").text("Create");
            $("#create-wallet-button").css('background-color', '#1428e5');
        }, 3000);
        WalletsTemplate = $('#all-wallets-template').html();
        TepmpalteDIV = "#display_all_wallets";
        var AllWallets = [];
        AllWallets[0] = {
            WalletAddress: createdAddress,
            WalletID: coinid,
            LastTrTime: "",
            LastTrValue: "",
        };
        ReplacedWalletsTemplate = replaceTemplate(AllWallets, "wallets", "AllWallets");
        $(TepmpalteDIV).prepend(ReplacedWalletsTemplate);
        if (document.getElementById("display_all_wallets").childElementCount >= 3) {
            $("#add_wallet_button").css("display", "none");
        }
        getBalance("wallets", 0);
    });

}

function loadSettings() {
    CoinID["Litecoin"] = 3;
    CoinID["EthClassic"] = 7;
    CoinID["Dogecoin"] = 8;
    CoinID["Bitcoin"] = 1;
    CoinID["Ether1"] = 6;
    CoinID["Ethereum"] = 4;
    CoinID["Ravecoin"] = 9;

    Coin2ID["BTC"] = 1;
    Coin2ID["ETC"] = 7;
    Coin2ID["LTC"] = 3;
    Coin2ID["DOGE"] = 8;
    Coin2ID["ETHO"] = 6;
    Coin2ID["ETH"] = 4;
    Coin2ID["RVN"] = 9;

    CoinById[3] = {
        B_ID: 3,
        ShortName: "LTC",
        Name: "Litecoin",
        ImgUrl: "http://" + HOST_IP + HOST_PORT + "/img/icon/wallets-icon/wallets-icon-litecoin.svg",
        Scale: 9,
        priceDelimiter: 1000000000,
        destination: 6,
        GasLimit: 1000,
        GasPrice: 0.00000001,
    };
    CoinById[7] = {
        B_ID: 7,
        ShortName: "ETC",
        Name: "EthClassic",
        ImgUrl: "http://" + HOST_IP + HOST_PORT + "/img/icon/wallets-icon/wallets-icon-ethclassic.svg",
        Scale: 9,
        priceDelimiter: 1000000000,
        destination: 6,
        GasLimit: 420,
        GasPrice: 0.00000001,
    };
    CoinById[8] = {
        B_ID: 8,
        ShortName: "DOGE",
        Name: "Dogecoin",
        ImgUrl: "http://" + HOST_IP + HOST_PORT + "/img/icon/wallets-icon/wallets-icon-dogecoin.svg",
        Scale: 9,
        priceDelimiter: 1000000000,
        destination: 6,
        GasLimit: 2100,
        GasPrice: 0.000001,
    };
    CoinById[1] = {
        B_ID: 1,
        ShortName: "BTC",
        Name: "Bitcoin",
        ImgUrl: "http://" + HOST_IP + HOST_PORT + "/img/icon/wallets-icon/wallets-icon-bitcoin.svg",
        Scale: 9,
        priceDelimiter: 1000000000,
        destination: 6,
        GasLimit: 1000,
        GasPrice: 0.00000001,
    };
    CoinById[6] = {
        B_ID: 6,
        ShortName: "ETHO",
        Name: "Ether1",
        ImgUrl: "http://" + HOST_IP + HOST_PORT + "/img/icon/wallets-icon/wallets-icon-ether1.svg",
        Scale: 9,
        priceDelimiter: 1000000000,
        destination: 6,
        GasLimit: 420,
        GasPrice: 0.000001,
    };
    CoinById[4] = {
        B_ID: 4,
        ShortName: "ETH",
        Name: "Ethereum",
        ImgUrl: "http://" + HOST_IP + HOST_PORT + "/img/icon/wallets-icon/wallets-icon-eth.svg",
        Scale: 9,
        priceDelimiter: 1000000000,
        destination: 6,
        GasLimit: 420,
        GasPrice: 0.000001,
    };
    CoinById[9] = {
        B_ID: 9,
        ShortName: "RVN",
        Name: "Ravecoin",
        ImgUrl: "http://" + HOST_IP + HOST_PORT + "/img/icon/wallets-icon/wallets-icon-ravecoin.svg",
        Scale: 9,
        priceDelimiter: 1000000000,
        destination: 6,
        GasLimit: 1000,
        GasPrice: 0.00000001,
    };
}

function selectWalletFrom(id, address) {
    $("#w_selected_name").empty();
    $("#w_selected_names").empty();
    $("#w_selected_name").append(CoinById[id].Name);
    $("#w_selected_names").append(CoinById[id].Name);
    $("#w_selected_img").attr("src", CoinById[id].ImgUrl);
    $("#w_selected_img2").attr("src", CoinById[id].ImgUrl);
    $("#wallet_addres").val(address);
    $("#wallet_addres").attr("CoinID", id);
    $("#sending_from").attr("CoinID", id);
    $("#sending_from").val(address);
    $("#coinShort2").empty();
    $("#coinShort").empty();
    $("#coinShort").append(CoinById[id].ShortName);
    $("#coinShort2").append(CoinById[id].ShortName);
    getBalance("transactions", id);

    if (document.getElementById("send_link").parentNode.querySelector(".is-active").id == "send_link") {
        displayQRCode("#w_addres_qr_send", address);
        DisplayWalletHistory(address, id, "send");
    } else {
        displayQRCode("#w_addres_qr_recive", address);
        DisplayWalletHistory(address, id, "receive");
    }

}

function filterWalletHistory(filter = "") {
    if (filter == "") {
        filter = 1;
    }

    var address = $("#wallet_addres").val();
    var coinID = $("#wallet_addres").attr("CoinID");
    if (coinID == null || coinID == 0) {
        var wallet = findGetParameter("wallet");;
        coinID = CoinID[wallet];
    }

    var formData = {
        data: "DisplayWalletHistory" + "," + session + "," + userId + "," + address + "," + filter + ",",
    };
    $('#transactionstemplateforrecive').empty();
    $('#transactionstemplateforsend').empty();
    $.post("http://" + HOST_IP + ":20001", formData, function (data) {

        var historyArr = [];
        if (data != "Not Found") {
            historyArr = data.split(" ");
        }

        historyArr.forEach(function (item, i, arr) {
            if (item != "") {
                var div = replaceSWalletHistory(item, coinID);
                var div2 = replaceSWalletHistory(item, coinID);
                $("#transactionstemplateforrecive").append(div);
                $("#transactionstemplateforsend").append(div2);
            }
        });
        var WalletsList = document.querySelector("#transactionstemplateforrecive");
        new SimpleBar(WalletsList).init();
        var WalletsList = document.querySelector("#transactionstemplateforsend");
        new SimpleBar(WalletsList).init();
    });
}

function DisplayWalletHistory(address, id, subpage) {
    var formData = {
        data: "DisplayWalletHistory" + "," + session + "," + userId + "," + address + "," + "0" + ",",
    };
    $.post("http://" + HOST_IP + ":20001", formData, function (data) {
        $('#transactionstemplateforrecive').empty();
        $("#transactionstemplateforsend").empty();
        $("#send-Tr").css("display", "none");
        $("#send-noTr").css("display", "block");
        $("#receive-Tr").css("display", "none");
        $("#receive-noTr").css("display", "block");
        var historyArr = [];
        if (data != "Not Found") {
            historyArr = data.split(" ");
        }
        if (subpage == "receive") {
            if ((historyArr.length - 1) > 0) {
                $("#receive-noTr").css("display", "none");
                $("#receive-Tr").css("display", "block");
            } else {
                $("#receive-Tr").css("display", "none");
                $("#receive-noTr").css("display", "block");
            }
        } else {
            if ((historyArr.length - 1) > 0) {
                $("#send-noTr").css("display", "none");
                $("#send-Tr").css("display", "block");
            } else {
                $("#send-Tr").css("display", "none");
                $("#send-noTr").css("display", "block");
            }
        }

        historyArr.forEach(function (item, i, arr) {
            if (item != "") {
                div = replaceSWalletHistory(item, id);
                if (subpage == "receive") {
                    document.getElementById('transactionstemplateforrecive').appendChild(div);
                }
                if (subpage == "send") {
                    document.getElementById('transactionstemplateforsend').appendChild(div);
                }
            }
        });
        if (subpage == "receive") {
            var WalletsList = document.querySelector("#transactionstemplateforrecive");
            new SimpleBar(WalletsList).init();
        } else {
            var WalletsList = document.querySelector("#transactionstemplateforsend");
            new SimpleBar(WalletsList).init();
        }

    });
}

function replaceSWalletHistory(item, id) {
    [TrValue, TrFrom, TrTo, TransactionHash, ConfirmedBlocks, InOut, TrTime] = item.split(",");
    var status = 0;
    if (InOut == "1") {
        status = "is-received";
    } else {
        status = "is-send";
    }
    let transactionsWalletTemplate = $("#transactions-page-card-template").html();
    let div = document.createElement("div");
    div.classList.add("transactions__card");
    div.style.height = "auto";
    div.style.maxHeight = "150px";
    inUSD = convertToDot(TrValue, CoinById[id].Scale, id) * CoinsPriceUSD[id];
    inUSD = parseFloat(inUSD).toFixed(2);
    transactionsWalletTemplate = transactionsWalletTemplate
        .replace(/{date}/g, timeConverter(TrTime))
        .replace(/{amount_in_wallet}/g, convertToDot(TrValue, CoinById[id].Scale, id))
        .replace(/{amount_in_wallet_currency}/g, CoinById[id].ShortName)
        .replace(/{amount_equivalent}/g, inUSD)
        .replace(/{amount_equivalent_currency}/g, "USD")
        .replace(/{transactions_id}/g, TransactionHash)
        .replace(/{transactions_from}/g, TrFrom)
        .replace(/{{confirm_blocks}}/g, ConfirmedBlocks)
        .replace(/{{destination}}/g, CoinById[id].destination)
        .replace(/{transactions_to}/g, TrTo);
    div.classList.add(status);
    div.innerHTML = transactionsWalletTemplate;
    return div
}

function timeConverter(UNIX_timestamp) {
    var a = new Date(UNIX_timestamp * 1000);
    var months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    var year = a.getFullYear();
    var month = months[a.getMonth()];
    var date = a.getDate();
    var hour = a.getHours();
    var min = a.getMinutes();
    var sec = a.getSeconds();
    var time = date + ' ' + month + ' ' + year + ' ' + hour + ':' + min;
    return time;
}

function findGetParameter(parameterName) {
    var result = null,
        tmp = [];
    location.search
        .substr(1)
        .split("&")
        .forEach(function (item) {
            tmp = item.split("=");
            if (tmp[0] === parameterName) result = decodeURIComponent(tmp[1]);
        });
    return result;
}

function displayAllHistory(page, filter = "") {
    if (filter == "") {
        filter = 1;
    }
    var formData = {
        data: "DisplayAllHistory" + "," + session + "," + userId + "," + filter + "," + ",",
    };
    $.post("http://" + HOST_IP + ":20001", formData, function (data) {
        $(TepmpalteDIV).empty();
        var ReplacedWalletsTemplate = "";
        var historyArr = [];
        historyArr = data.split(" ");
        var countTr = 0;
        if (historyArr != "") {
            countTr = historyArr.length - 1;
        }

        $("#all_tr_desctop").empty();
        $("#all_tr_mobile").empty();
        $("#all_tr_desctop").append(countTr);
        $("#all_tr_mobile").append(countTr);

        if (page == "dashboard") {
            if (countTr >= 1) {
                $("#dashboard-Tr").css("display", "block");
                $("#r_ballance-list").css("display", "block");

            } else {
                $("#dashboard-noTr").css("display", "block");
                $("#no-balance").css("display", "block");
            }
        }
        if (page == "history") {
            if (countTr >= 1) {
                $("#dashboard-noTr").css("display", "none");
                $("#all_history_template").css("display", "block");
            } else {
                $("#all_history_template").css("display", "none");
                $("#dashboard-noTr").css("display", "block");
            }
        }

        historyArr.forEach(function (item, i, arr) {
            if (item != "") {
                // [TrValue, TrFrom, TrTo, TransactionHash, ConfirmedBlocks, InOut, TrTime] = item.split(",");
                replData = item.split(",");
                if (page == "dashboard") {
                    ReplacedWalletsTemplate = replaceTemplate(replData, page, "");
                    var status = "is-received";
                    if (replData[5] == "1") {
                        status = "is-received";
                    } else {
                        status = "is-send";
                    }

                    var article = document.createElement("article");
                    article.classList.add("transactions__card");
                    article.classList.add(status);
                    article.innerHTML = ReplacedWalletsTemplate;

                    document.getElementById("r_transactions-content-dashboard").appendChild(article);

                } else {
                    ReplacedWalletsTemplate = replaceTemplate(replData, page, "");
                    $(TepmpalteDIV).append(ReplacedWalletsTemplate);
                }

            }
        });
        if (page == "dashboard") {
            var WalletsList = document.querySelector("#r_transactions-content-dashboard");
            new SimpleBar(WalletsList).init();
        }
    });
}

function displayAllWallets(page, filter) {
    if (filter == "") {
        filter = 1;
    }
    var formData = {};
    var AllWallets = [];
    var ReplacedWalletsTemplate = "";
    if (page == "wallets" || page == "main") {
        formData = {
            data: "DisplayAllWallets" + "," + session + "," + userId + "," + filter + "," + ",",
        };
    }
    if (page == "transactions") {
        formData = {
            data: "DisplayAllWalletsTR" + "," + session + "," + userId + "," + filter + "," + ",",
        };
    }

    $.post("http://" + HOST_IP + ":20001", formData, function (data) {
        if (data == "No Wallets") {
            data = "";
        }
        if (page == "main") {
            $("#activeWallets").empty();
            $("#activeWallets").append(0);
        }
        if (page == "transactions") {
            $('#transactions-template-for-recive').empty();
        }
        wallets = data.split(" ");
        wallets.forEach(function (item, i, arr) {
            if (item.length > 1) {
                WalletsData = item.split(",");
                AllWallets[i] = {
                    WalletAddress: WalletsData[0],
                    WalletID: WalletsData[1],
                    LastTrTime: WalletsData[2],
                    LastTrValue: WalletsData[3],
                }
            }
        });
        ReplacedWalletsTemplate = "";
        if (wallets != "") {
            if (page == "wallets") {
                ReplacedWalletsTemplate = replaceTemplate(AllWallets, page, "AllWallets");
                if (AllWallets.length >= 3) {
                    //$(TepmpalteDIV).empty();
                    $("#add_wallet_button").css("display", "none");
                } else {
                    $("#add_wallet_button").css("display", "flex");
                }
                $(TepmpalteDIV).prepend(ReplacedWalletsTemplate);
                getBalance(page, 0);
                ReplacedWalletsTemplate = ""
            }
            if (page == "main") {
                var test = "";
                var Template = $('#wallet_list').html();
                $("#activeWallets").empty();
                $("#activeWallets").append(wallets.length - 1);
                changeOwncarusel();

                AllWallets.forEach(function (item, i, arr) {
                    B_ID[i] = CoinById[item.WalletID].B_ID;
                    CoinType = CoinById[item.WalletID].ShortName;
                    imgURL = CoinById[item.WalletID].ImgUrl;
                    WalletTitle = CoinById[item.WalletID].Name;

                    test = Template.replace(/{{wallet_title}}/g, WalletTitle)
                        // .replace(/{coind}/g, CoinById[item.WalletID].B_ID)
                        .replace(/wallet_main_balance/g, "wallet_main_balance" + CoinById[item.WalletID].B_ID)
                        .replace(/{{wallet_img_url}}/g, imgURL)
                        .replace(/{coind}/g, CoinById[item.WalletID].B_ID)
                        .replace(/balance_in_USD/g, "balance_in_USD" + CoinById[item.WalletID].B_ID)
                        .replace(/{{last_trunsaction_date}}/g, new Date(item.LastTrTime * 1000))
                        .replace(/{{trunsaction_value}}/g, parseFloat(item.LastTrValue / CoinById[item.WalletID].priceDelimiter))
                        .replace(/{{coin_type}}/g, CoinType.toLowerCase());

                    $('.my-wallets__carousel').slick('slickAdd', test, true);
                    focusCarouselItem();
                });

                $('.my-wallets__carousel').slick("setPosition");

                setTimeout(function () {
                    getBalance(page, 0);
                }, 1000);

                ReplacedWalletsTemplate = "";
            }
            if (page == "transactions") {
                ReplacedWalletsTemplate = replaceTemplate(AllWallets, page, "AllWallets");
                $(TepmpalteDIV).append(ReplacedWalletsTemplate);
                $("#transactions_send_w_add").append(ReplacedWalletsTemplate);
                ReplacedWalletsTemplate = "";
            }
        }
    });
}

function changeOwncarusel() {
    let walletsCarousel = $(".my-wallets__carousel");
    if (walletsCarousel) {
        walletsCarousel.slick({
            loop: false,
            variableWidth: true,
            infinite: false,
            focusOnSelect: false,
            mobileFirst: true,
            swipeToSlide: true,
            arrows: false,
            responsive: [{
                    loop: false,
                    breakpoint: 375,
                    settings: {
                        loop: false,
                        arrows: false,
                        infinite: false,
                        swipeToSlide: true,
                    }
                },
                {
                    loop: false,
                    breakpoint: 550,
                    settings: {
                        loop: false,
                        arrows: false,
                        infinite: false,
                        swipeToSlide: true,
                    }
                },
                {
                    loop: false,
                    breakpoint: 992,
                    settings: {
                        loop: false,
                        arrows: true,
                        infinite: false,
                        swipeToSlide: true,
                    }
                },
            ]
        });
    }
}

function focusCarouselItem() {
    let carouselItem = document.querySelectorAll(".my-wallets__wrapper");

    if (carouselItem) {
        for (let i = 0; i < carouselItem.length; i++) {
            carouselItem[i].addEventListener("click", function () {
                let current = document.getElementsByClassName("style-is-focus");
                // If there's no active class
                if (current.length > 0) {
                    current[0].className = current[0].className.replace(
                        " style-is-focus",
                        ""
                    );
                }
                // Add the active class to the current/clicked button
                this.children[0].className += " style-is-focus";

                return false;
            });
        }
    }
}

function replaceTemplate(dataW, page, TemplateType) {
    if (page == "wallets" || page == "main") {
        var ReplacedWalletsTemplate = "";
        var WalletBalance = 0;
        dataW.forEach(function (item, i, arr) {
            B_ID[i] = CoinById[item.WalletID].B_ID;
            CoinType = CoinById[item.WalletID].ShortName;
            imgURL = CoinById[item.WalletID].ImgUrl;
            WalletTitle = CoinById[item.WalletID].Name;

            if (page == "wallets") {
                WalletsTemplate = $('#all-wallets-template').html();
            }
            ReplacedWalletsTemplate = WalletsTemplate.replace(/{{wallet_title}}/g, WalletTitle)
                .replace(/{coind}/g, CoinById[item.WalletID].B_ID)
                .replace(/{{Wallet_Name}}/g, CoinById[item.WalletID].Name)
                .replace(/{{WalletAddress}}/g, item.WalletAddress)
                .replace(/wallet_main_balance/g, "wallet_main_balance" + CoinById[item.WalletID].B_ID)
                .replace(/{{wallet_img_url}}/g, imgURL)
                .replace(/{{last_trunsaction_date}}/g, new Date(item.LastTrTime * 1000))
                .replace(/{{trunsaction_value}}/g, parseFloat(item.LastTrValue / CoinById[item.WalletID].priceDelimiter))
                .replace(/{{coin_type}}/g, CoinType.toLowerCase()) + ReplacedWalletsTemplate;
        });
        return ReplacedWalletsTemplate;
    }
    if (page == "transactions") {
        var ReplacedWalletsTemplate = "";
        var WalletBalance = 0;
        dataW.forEach(function (item, i, arr) {
            B_ID[i] = CoinById[item.WalletID].B_ID;
            CoinType = CoinById[item.WalletID].ShortName;
            imgURL = CoinById[item.WalletID].ImgUrl;
            WalletTitle = CoinById[item.WalletID].Name;

            ReplacedWalletsTemplate = WalletsTemplate.replace(/{{wallet_title}}/g, WalletTitle)
                .replace(/{{wallet_img_url}}/g, imgURL)
                .replace(/{{wallet_address}}/g, item.WalletAddress)
                .replace(/{{wallet_by_id}}/g, CoinById[item.WalletID].B_ID)
                .replace(/wallet_list_/g, "wallet_list_" + CoinById[item.WalletID].B_ID) + ReplacedWalletsTemplate;
        });
        return ReplacedWalletsTemplate;
    }
    if (page == "history" || page == "dashboard") {
        //[TrValue, TrFrom, TrTo, TransactionHash, ConfirmedBlocks, InOut, TrTime, CoinID]
        var status = "is-received ";
        var PDC_status = "is-completed";
        var textStatus = "Completed";
        if (dataW[5] == "1") {
            status = "is-received ";
        } else {
            status = "is-send ";
        }

        var CoinId = dataW[7];
        if (dataW[4] >= CoinById[CoinId].destination) {
            PDC_status = "is-completed ";
            textStatus = "Completed";
        } else {
            textStatus = "Pending";
            PDC_status = "is-pending ";
        }
        inUSD = convertToDot(dataW[0], CoinById[CoinId].Scale, CoinId) * CoinsPriceUSD[CoinId];
        inUSD = parseFloat(inUSD).toFixed(2);
        ReplacedWalletsTemplate = WalletsTemplate
            .replace(/{date}/g, timeConverter(dataW[6]))
            .replace(/{amount_in_wallet}/g, convertToDot(dataW[0], CoinById[CoinId].Scale, CoinId))
            .replace(/{amount_in_wallet_currency}/g, CoinById[CoinId].ShortName)
            .replace(/{amount_equivalent}/g, inUSD)
            .replace(/{amount_equivalent_currency}/g, "USD")
            .replace(/{transactions_id}/g, dataW[3])
            .replace(/{transactions_from}/g, dataW[1])
            .replace(/{{confirm_blocks}}/g, dataW[4])
            .replace(/{{is_RSP_type}} /g, status)
            .replace(/{{img_URL}}/g, CoinById[CoinId].ImgUrl)
            .replace(/{{destination}}/g, CoinById[CoinId].destination)
            .replace(/{{textStatus}}/g, textStatus)
            .replace(/{{PDC_status}}/g, PDC_status)
            .replace(/{transactions_to}/g, dataW[2]);
        return ReplacedWalletsTemplate;
    }
}

function ScaleN(digit) {
    return Math.round(Math.log(digit / Math.log(10)));
}

function getBalance(page, id) {
    if (page == "wallets" || page == "main") {
        B_ID.forEach(function (coinId, i, arr) {
            param = {
                data: 'balanceW,' + session + ',' + userId + ',' + coinId + ',,,'
            };

            $.post("http://" + HOST_IP + ":10000", param, function (data) {
                [balance, balanceUSD] = data.split(" ");
                balanceUSD = parseFloat(balanceUSD / CoinById[coinId].priceDelimiter).toFixed(2);
                balanceCMP = balance;
                balance = convertToDot(balance, CoinById[coinId].Scale, coinId);
                var j = 0;
                var attr_w1 = $('[data-attr="w1allet_b1alance"]');
                var attr_w2 = $('[data-attr="w2allet_b2alance"]');

                flag = false;

                for (j = 0; j < 150; j++) {
                    if (attr_w1[j]) {
                        attr = "'" + attr_w1[j].attributes.coind.value + "'";
                        id = "'" + coinId + "'";
                        if (attr == id && attr_w1[j].innerHTML == "") {

                            flag = true;
                            break;
                        }
                    }
                }
                if (page == "wallets") {
                    divID = "#wallet_" + coinId;
                    if (balanceCMP == 0 || balanceCMP == "0") {
                        $(divID).addClass("no-balance");
                    } else {
                        $(divID).addClass("balance-is-there");
                    }
                }
                if (flag) {
                    if (page == "main") {
                        attr_w2[j].innerHTML = balanceUSD + "<span>usd</span>";
                    } else {
                        attr_w2[j].innerHTML = balanceUSD + '<span class="my-wallet__sub"> usd</span>';
                    }
                    attr_w1[j].innerHTML = balance + "<span class='my-wallet__sub'>" + CoinById[coinId].ShortName + "</span>";
                }

            });

        });

    }
    if (page == "transactions") {
        param = {
            data: 'balanceW,' + session + ',' + userId + ',' + id + ',,,'
        };
        $.post("http://" + HOST_IP + ":10000", param, function (data) {
            [balance, balanceUSD] = data.split(" ");
            balanceUSD = parseFloat(balanceUSD / CoinById[id].priceDelimiter).toFixed(2);
            balance = convertToDot(balance, CoinById[id].Scale, id);

            BalanceID = "#wallet_balance";
            BalanceID2 = "#wallet_balance2";
            USDBalanceID = "#balance_in_USD";
            USDBalanceID2 = "#balance_in_USD2";

            $(BalanceID).empty();
            $(BalanceID2).empty();
            $(BalanceID).append(balance);
            $(BalanceID2).append(balance);
            $(USDBalanceID).empty();
            $(USDBalanceID2).empty();
            $(USDBalanceID).append(balanceUSD + " USD");
            $(USDBalanceID2).append(balanceUSD + " USD");

            if ($("#transaction_curency")) {
                $("#transaction_curency").empty();
                $("#transaction_curency2").empty();
                $("#transaction_curency_f").empty();
                $("#transaction_curency_f2").empty();
                $("#transaction_curency_u").empty();
                $("#transaction_curency_u2").empty();
                $("#transaction_curency_f").append(CoinById[id].Name + " balance: ");
                $("#transaction_curency_f2").append(CoinById[id].Name + " balance: ");
                $("#transaction_curency_u").append("usd");
                $("#transaction_curency_u2").append("usd");
                $("#transaction_curency").append(CoinById[id].ShortName);
                $("#transaction_curency2").append(CoinById[id].ShortName);
            }
        });
    }
}

function convertFromDot(convData, scale) {
    convData = String(convData);
    convData = convData.split('.');
    if (convData[1] == null) {
        convData[1] = "0";
    }
    convData[1] = convData[1].padEnd(scale, '0');
    result = convData[0] + convData[1];
    return result;
}

function convertToDot(convData, scale, coinId) {
    convData = String(convData);
    convData = convData.padStart(scale + 1, '0');
    var rplc = "(.{" + scale + "}$)";
    var re = new RegExp(rplc, "g");
    result = convData.replace(re, ".$1");
    result = result.replace(CoinById[coinId].Delimetr, "");
    return result
}

function getOverviewData() {
    //get all balance in usd
    //get all balance in btc
    //get all balance in Uero
    param = {
        data: "CoutTransactions" + "," + session + "," + userId + "," + "," + "," + ",",
    };
    $.post("http://" + HOST_IP + ":20001", param, function (data) {
        if (data != "") {
            $("#completedTtransaction").append(data);
        } else {
            $("#completedTtransaction").append(0);
        }
    });
    param = {
        data: 'allbalanceW,' + session + ',' + userId + ',' + "," + ',,,'
    };

    $.post("http://" + HOST_IP + ":10000", param, function (data) {
        [BalanceIn, BalanceInPData] = data.split(" ");
        [BalanceInUSD, BalanceInEUR, BalanceInBTC] = BalanceIn.split(",");

        $("#OverviewBalanceUSD").html('<span class="current-balance__sup">USD</span>');
        $("#OverviewBalanceEUR").html('<span class="main-currency__currency">eur</span>');
        $("#OverviewBalanceBTC").html('<span class="current-balance__sup">BTC</span>');
        $("#OverviewBalanceUSD").prepend(parseFloat(BalanceInUSD).toFixed(2));
        $("#OverviewBalanceEUR").prepend(parseFloat(BalanceInEUR).toFixed(2));
        $("#OverviewBalanceBTC").prepend(parseFloat(BalanceInBTC).toFixed(5));
        //
        BalanceInPercent = BalanceInPData.split("|");
        BalanceInPercent.forEach(function (itm, i, arr) {
            Pdata = itm.split(",");
            if (Pdata != "") {
                value = parseFloat(Pdata[1]).toFixed(2);
                let li = document.createElement("li");
                li.classList.add("ballance__item");
                let ballanceTemplate = jQuery("#r_ballance-template").html();
                if (ballanceTemplate) {
                    ballanceTemplate = ballanceTemplate
                        .replace(/{percentage}/g, value)
                        .replace(/{ballanceName}/g, CoinById[Pdata[0]].Name);
                    li.innerHTML = ballanceTemplate;
                    document.getElementById("r_ballance-list").appendChild(li);
                }
            }
        });
    });
}

function coutFee() {
    var countGas = 0.0;
    $("#trnFee").empty();

    var coinID = $("#wallet_addres").attr("CoinID");
    if (coinID == null || coinID == 0) {
        var wallet = findGetParameter("wallet");
        coinID = CoinID[wallet];
    }

    countGas = CoinById[coinID].GasLimit * CoinById[coinID].GasPrice; // * ourFee
    var value = $("#send_amout_c").val();
    var total = parseFloat(value) + parseFloat(countGas);
    var InUSD = value * CoinsPriceUSD[coinID];

    GasInUSD = parseFloat(countGas) * CoinsPriceUSD[coinID];
    InUSDPlusFee = parseFloat(InUSD) + parseFloat(GasInUSD);

    InUSD = parseFloat(InUSD).toFixed(2);
    GasInUSD = parseFloat(GasInUSD).toFixed(2);
    InUSDPlusFee = parseFloat(InUSDPlusFee).toFixed(2);

    $("#totalInUSD").empty();
    $("#inUSDWithFee").empty();
    $("#gasInUSD").empty();
    $("#totalInUSD").append(parseFloat(total).toFixed(10) + " " + CoinById[coinID].ShortName);
    $("#gasInUSD").append(GasInUSD + " USD");
    $("#inUSDWithFee").append(InUSDPlusFee + " USD");
    $("#trnFee").append(countGas.toFixed(8) + " " + CoinById[coinID].ShortName);

}

function loadWInside(e) {
    window.location.replace(e.id);
}

function QuickSend() {
    var sendCoinsTo = $("#QS_sndTO").val().replace(/ /g, "");
    var value = $("#QS_amount").val();
    var flag = false;
    $("#QS_sndTO").css("border", "1px solid transparent");

    if (value == null) {
        value = $("#send_amout_c").val();
        flag = true;
        $("#send_amout_c").css("border", "1px solid transparent");
    } else {
        $("#QS_amount").css("border", "1px solid transparent");
    }

    $("#modal__desc").empty();
    $("#modal_TR_desc").empty();

    if (value <= 0) {
        $("#attention3").addClass("active");
        $("#modal_overlay").addClass("active");
        $("#modal__desc").append("Error: Set the value up!");
        return
    }

    var wallet = findGetParameter("wallet");

    if (wallet == null) {
        wallet = $("#quick-sender-select").val()
    }

    var coinID = CoinID[wallet];
    value = convertFromDot(value, CoinById[coinID].Scale);

    if (value != 0 && coinid != 0) {
        if (value.length > 0 && coinid > 0 && sendCoinsTo.length > 10) {
            param = {
                data: "QuickSendCoins" + "," + session + "," + userId + "," + "," + "," + ",",
                sndTo: sendCoinsTo,
                CoinID: coinID,
                value: value,
            };
            $.post("http://" + HOST_IP + ":20001", param, function (data) {
                [msg, flag] = data.split(",");
                if (flag == "1") {
                    $("#almost-done").addClass("active");
                    $("#modal_overlay").addClass("active");
                    $("#modal_TR_desc").append(msg);
                } else {
                    $("#attention3").addClass("active");
                    $("#modal_overlay").addClass("active");
                    $("#modal__desc").append(msg);
                }
            });
        } else {
            $("#QS_sndTO").css("border", "1px solid rgb(193, 30, 15)");
            $("#attention3").addClass("active");
            $("#modal_overlay").addClass("active");
            $("#modal__desc").append("Error: Wrong Address!");
        }
    } else {
        if (flag == false) {
            $("#QS_amount").css("border", "1px solid rgb(193, 30, 15)");
        } else {
            $("#send_amout_c").css("border", "1px solid rgb(193, 30, 15)");
        }
        $("#attention3").addClass("active");
        $("#modal_overlay").addClass("active");
        $("#modal__desc").append("Error: Select Coin or set the value up!");
    }
}

function sendMixer(e, eType) {
    $("#TrError").empty();
    $("#TrError").css("display", "none");
    $("#TrDone").css("display", "none");

    if (eType == "open") {
        $("#mixerSend").css("display", "block");
        $("#closeBTN").css("display", "none");
        $("#totalInUSD").empty();
        $("#inUSDWithFee").empty();
        $("#gasInUSD").empty();
        $("#trnFee").empty();


        $("#inUSDWithFee").append("0.0 USD");
        $("#gasInUSD").append("0.0 USD");
        $("#trnFee").append("0.0 BTC");

        var SendCoinsTo = e.parentNode.querySelector(".contact__wallet-id").innerHTML;
        var CoinID = e.parentNode.querySelector(".contact__icon").getAttribute("coinid");

        $("#totalInUSD").append("0.0 " + CoinById[CoinID].ShortName);

        $("#CoinType").empty();
        $("#SendCoinsTo").val(SendCoinsTo.replace(/ /g, ""));
        $("#CoinType").append(CoinById[CoinID].ShortName);
    }

    if (eType == "send") {
        $("#SendAmount").css("border", "1px solid transparent");
        $("#SendCoinsTo").css("border", "1px solid transparent");

        var value = $("#SendAmount").val();
        if (value <= 0) {
            $("#TrError").css("display", "block");
            $("#TrError").append("Error: Set the value up!");
            return
        }
        var Coin = $("#CoinType").text();

        var coinid = Coin2ID[Coin];
        value = convertFromDot(value, CoinById[coinid].Scale);

        var sendCoinsTo = $("#SendCoinsTo").val();
        sendCoinsTo = sendCoinsTo.replace(/ /g, "");
        if (value != 0 && coinid != 0) {

            if (value.length > 0 && coinid > 0 && sendCoinsTo.length > 10) {
                param = {
                    data: "QuickSendCoins" + "," + session + "," + userId + "," + "," + "," + ",",
                    sndTo: sendCoinsTo,
                    CoinID: coinid,
                    value: value,
                };
                $.post("http://" + HOST_IP + ":20001", param, function (data) {
                    [msg, flag] = data.split(",");
                    if (flag == "1") {
                        $("#TrDone").css("display", "block");
                        $("#mixerSend").css("display", "none");
                        $("#closeBTN").css("display", "block");
                        setTimeout(function () {
                        },3000);
                    } else {
                        $("#TrError").css("display", "block");
                        $("#TrError").append(msg);
                    }
                });
            } else {
                if (sendCoinsTo.length < 10) {
                    $("#SendCoinsTo").css("border", "1px solid rgb(193, 30, 15)");
                }
                $("#TrError").css("display", "block");
                $("#TrError").append("Error: Wrong Address!");
            }
        } else {
            if (value == 0) {
                $("#SendAmount").css("border", "1px solid rgb(193, 30, 15)");
            }

            $("#TrError").css("display", "block");
            $("#TrError").append("Error: Select Coin or set the value up!");
        }
    }
    if (eType == "countFee") {
        var countGas = 0.0;
        $("#trnFee").empty();

        var Coin = $("#CoinType").text();
        var coinID = Coin2ID[Coin];

        countGas = CoinById[coinID].GasLimit * CoinById[coinID].GasPrice; // * ourFee

        var value = $("#SendAmount").val();
        var total = parseFloat(value) + parseFloat(countGas);
        var InUSD = value * CoinsPriceUSD[coinID];

        GasInUSD = parseFloat(countGas) * CoinsPriceUSD[coinID];
        InUSDPlusFee = parseFloat(InUSD) + parseFloat(GasInUSD);

        InUSD = parseFloat(InUSD).toFixed(2);
        GasInUSD = parseFloat(GasInUSD).toFixed(2);
        InUSDPlusFee = parseFloat(InUSDPlusFee).toFixed(2);

        $("#totalInUSD").empty();
        $("#inUSDWithFee").empty();
        $("#gasInUSD").empty();
        $("#totalInUSD").append(parseFloat(total).toFixed(10) + " " + CoinById[coinID].ShortName);
        $("#gasInUSD").append(GasInUSD + " USD");
        $("#inUSDWithFee").append(InUSDPlusFee + " USD");
        $("#trnFee").append(countGas + " " + CoinById[coinID].ShortName);
    }
}

function SendCoinsTo() {
    var value = 0.0;
    var sendCoinsTo = "";
    var sendFrom = "";
    var coinid = 0;

    $("#modal_TR_desc").empty();
    $("#modal__desc").empty();

    sendCoinsTo = $("#sendCoinsTo").val();
    sendFrom = $("#sending_from").val();
    coinid = $("#sending_from").attr("CoinID");
    value = $("#send_amout_c").val();

    $("#sendCoinsTo").css("border", "1px solid transparent");
    $("#send_amout_c").css("border", "1px solid transparent");
    $("#sending_from").css("border", "1px solid transparent");

    if (value <= 0) {
        $("#attention3").addClass("active");
        $("#modal_overlay").addClass("active");
        $("#modal__desc").append("Error: Set the value up!");
        return
    }

    if (value != 0 && coinid != 0) {

        value = convertFromDot(value, CoinById[coinid].Scale);

        if (value.length > 0 && coinid.length > 0 && sendCoinsTo.length > 10 && sendFrom.length > 10) {
            param = {
                data: "SendCoinsTo" + "," + session + "," + userId + "," + "," + "," + ",",
                sndFrom: sendFrom,
                sndTo: sendCoinsTo,
                CoinID: coinid,
                value: value,
            };
            $.post("http://" + HOST_IP + ":20001", param, function (data) {
                [msg, flag] = data.split(",");
                if (flag == "1") {
                    $("#almost-done").addClass("active");
                    $("#modal_overlay").addClass("active");
                    $("#modal_TR_desc").append(msg);
                } else {
                    $("#attention3").addClass("active");
                    $("#modal_overlay").addClass("active");
                    $("#modal__desc").append(msg);
                }
            });
        } else {
            if (value.length <= 0) {
                $("#send_amout_c").css("border", "1px solid rgb(193, 30, 15)");
            }
            if (sendFrom.length <= 10) {
                $("#sending_from").css("border", "1px solid rgb(193, 30, 15)");
            }
            if (sendCoinsTo.length <= 10) {
                $("#sendCoinsTo").css("border", "1px solid rgb(193, 30, 15)");
            }
            $("#attention3").addClass("active");
            $("#modal_overlay").addClass("active");
            $("#modal__desc").append("Error: Wrong Address!");
        }
    } else {
        if (value == 0) {
            $("#send_amout_c").css("border", "1px solid rgb(193, 30, 15)")
            if (coinid == 0) {
                $("#sending_from").css("border", "1px solid rgb(193, 30, 15)");
                $("#sendCoinsTo").css("border", "1px solid rgb(193, 30, 15)");
            }
        } else {
            $("#sending_from").css("border", "1px solid rgb(193, 30, 15)");
            $("#send_amout_c").css("border", "1px solid rgb(193, 30, 15)");
            $("#sendCoinsTo").css("border", "1px solid rgb(193, 30, 15)");
        }

        $("#attention3").addClass("active");
        $("#modal_overlay").addClass("active");
        $("#modal__desc").append("Error: Select Coin or set the value up!");
    }
}

//convert coin to coin or coin to monay
function covertTo() {
    param = {
        data: 'AllToUSD,' + session + ',' + userId + ',' + "," + ',,,'
    };
    $.post("http://" + HOST_IP + ":10000", param, function (data) {
        response = data.split(" ");
        //CoinsPriceUSD
        response.forEach(function (coinsData, i, arr) {
            if (coinsData != "") {
                [coinid, coinValue] = coinsData.split(",");
                CoinsPriceUSD[coinid] = coinValue;
            }
        });
    });
}

function loadMarkets() {
    var marketsContainer = "#all_markets_container";
    var marketsTemplate = $("#markets_template").html();

    CoinsPriceUSD.forEach(function (coinsData, i, arr) {
        var div = marketsTemplate;
        var Pair = CoinById[i].ShortName + "/USD";
        div = div.replace(/{{value}}/g, parseFloat(coinsData).toFixed(5))
            .replace(/{{pair}}/g, Pair);

        $(marketsContainer).append(div);
    });
}

function displayCharts() {

    // widgetOptions = {
    //     debug: true,
    //     symbol: "BTC" + '/' + "USD",
    //     datafeed: Datafeed, // our datafeed object
    //     interval: '1',
    //     container_id: 'chart',
    //     library_path: '/charting_library/charting_library/',
    //     locale: 'en',
    //     timezone: 'Europe/Athens',
    //     disabled_features: ['use_localstorage_for_settings', 'volume_force_overlay', 'move_logo_to_main_pane', 'timeframes_toolbar'],
    //     enabled_features: ['support_multicharts', 'chart_property_page_scales'],
    //     client_id: 'test',
    //     user_id: 'public_user_id',
    //     fullscreen: false,
    //     autosize: true,
    //     overrides: {
    //         "paneProperties.background": chart_bg_color, //dark background
    //         "paneProperties.vertGridProperties.color": "#d8d8d8",
    //         "paneProperties.horzGridProperties.color": "#d8d8d8",
    //         "symbolWatermarkProperties.transparency": 0,
    //         "scalesProperties.textColor": chart_text_color,
    //         "mainSeriesProperties.candleStyle.wickUpColor": '#39b54a',
    //         "mainSeriesProperties.candleStyle.wickDownColor": '#7f323f',
    //         "volumePaneSize": "small",
    //     }
    // };
    //  window.tvWidget = new TradingView.widget(widgetOptions);

    d = new Date();
    addBar = {
        time: d.setSeconds(0, 0),
        open: 1,
        high: 1,
        low: 1,
        close: 1,
        volume: 1,
    };
    feed.unshift(addBar);

    lastBar = {
        close: close
    };

    run.shift().call();

    setTimeout(function () {

        subscription({
            time: d.setSeconds(0, 0),
            open: 1,
            high: 1,
            low: 1,
            close: 1,
            volume: 1
        })
    }, 10000);

}

function displayQRCode(ID, text) {
    $(ID).empty();
    $(ID).data("qr", {
        "render": "image",
        "size": 256,
        "text": text
    }).qrcode($(ID).data("qr"))

}

function displayContacts(IsData = "") {
    var Template = $("#contacts_all_w_template").html();
    var appendDiv = "#add_newContact_form";
    var appendDivWallet = ".contacts__item.contact.no-delete";

    if (IsData != "") {
        var arrData = IsData.split("|");
        arrData.forEach(function (item, i, arr) {
            [UserFIO, About, CoinId, WalletAddr, ImgUrl] = item.split(",")
            if (CoinId != null && CoinId != 0) {
                var tmpData = Template.replace(/{{FIO}}/g, UserFIO)
                    .replace(/{{About}}/g, About)
                    .replace(/{{WAddress}}/g, WalletAddr)
                    .replace(/{{CoinURL}}/g, CoinById[CoinId].ImgUrl)
                    .replace(/{{ImgURL}}/g, "http://" + HOST_IP + "/ConatctImages/" + ImgUrl);
                $(appendDiv).prepend(tmpData);
                addmodals()
            }
        });
    } else {
        formData = {
            data: "DisplayAllContacts" + "," + session + "," + userId + "," + "," + "," + ",",
        };
        $.post("http://" + HOST_IP + ":21001", formData, function (data) {
            var arrData = data.split("|");

            var lenData = arrData.length - 1;
            var ij = 0;
            if (lenData == 0) {
                $("#no-Contacts").css("display", "flex");
            }
            arrData.forEach(function (item, i, arr) {
                ij++;
                [UserFIO, About, CoinId, WalletAddr, ImgUrl] = item.split(",")
                if (CoinId != null && CoinId != 0) {
                    if (ImgUrl.length < 3) {
                        ImgUrl = "./img/ava.png"
                    } else {
                        ImgUrl = "http://" + HOST_IP + "/ConatctImages/" + ImgUrl
                    }
                    var tmpData = Template.replace(/{{FIO}}/g, UserFIO)
                        .replace(/{{About}}/g, About)
                        .replace(/{{WAddress}}/g, WalletAddr)
                        .replace(/{{CoinURL}}/g, CoinById[CoinId].ImgUrl)
                        .replace(/{{CoinID}}/g, CoinId)
                        .replace(/{{ImgURL}}/g, ImgUrl);
                    $(appendDivWallet).after(tmpData);
                }
                addmodals();
            });

        });
    }
}

function addmodals() {
    {
        let overlayModal = document.querySelector(".modal-overlay");
        let modals = document.querySelectorAll(".modal");
        let triggerModal = document.querySelectorAll(".modal-trigger");
        let closersModal = document.querySelectorAll(".modal__close");
        let modalsLen = modals.length;
        if (triggerModal.length !== 0) {
            for (let i = 0; i < modalsLen; i++) {
                if (triggerModal[i] != null) {
                    triggerModal[i].addEventListener("click", getId, false);
                }
                if (closersModal[i] != null) {
                    closersModal[i].addEventListener("click", close, false);
                }
                if (overlayModal[i] != null) {
                    overlayModal.addEventListener("click", closeOverlay, false);
                }
            }

            function getId(event) {
                event.preventDefault();
                let self = this;
                // get the value of the data-modal attribute from the button
                let modalId = self.dataset.modal;
                if (modalId == null) {
                    return
                }
                let len = modalId.length;
                // remove the '#' from the string
                let modalIdTrimmed = modalId.substring(1, len);
                // select the modal we want to activate
                let modal = document.getElementById(modalIdTrimmed);

                overlayModal.classList.add("active");
                modal.classList.add("active");

                if ($('#created_wallet_address') != null) {
                    $('#created_wallet_address').empty();
                    $('#created_wallet_address').val("...");

                }
            }

            function close(event) {
                event.preventDefault();
                let self = this;
                let modalActive = document.querySelector(".modal.active");
                if (modalActive) {
                    modalActive.classList.remove("active");
                    overlayModal.classList.remove("active");
                }
            }

            function closeOverlay(event) {
                let overlayModalActive = document.querySelector(
                    ".modal-overlay.active"
                );
                let modalActive = document.querySelector(".modal.active");
                if (overlayModalActive) {
                    if (overlayModal === event.target) {
                        modalActive.classList.remove("active");
                        overlayModalActive.classList.remove("active");
                    }
                }
            }
        }
    }
}

function saveImg() {
    document.querySelector('input[type="file"]').addEventListener('change', function () {
        if (this.files && this.files[0]) {
            var src = URL.createObjectURL(this.files[0]); // set src to blob url
            var CurrentPage = window.location.pathname;
            CurrentPage = CurrentPage.replace("/", "");
            $("#EnableUploadError").removeClass("warning-error");

            if (this.files[0].type != "image/jpg" && this.files[0].type != "image/jpeg" && this.files[0].type != "image/png") {
                $("#EnableUploadError").addClass("warning-error");
                return;
            }
            if (CurrentPage == "contacts.html") {
                $("#contactPic").attr("src", src);
                var fd = new FormData();
                fd.append('uploadFile', this.files[0]);
                fd.append('data', userId);
                $.ajax({
                    url: "http://" + HOST_IP + ":21002/upload",
                    type: 'post',
                    data: fd,
                    contentType: false,
                    processData: false,
                    success: function (response) {
                        console.log(response);
                    },
                    error: function (response) { // error callback
                        console.log(response);
                        alert("Error");
                    }
                });
            }
        }
    });
}

function addNewContact() {
    var UserFIO = $("#UserFIO").val();
    var About = $("#UserAbout").val();
    var CoinId = $("#UserCoinId").val();
    var WalletAddr = $("#WalletAddr").val();
    var UserID = userId;

    $("#ContactCreated").css("display", "none");
    $("#EnableFioError").removeClass("warning-error");
    $("#FioError").empty();
    $("#WalletAddr").css("border", "1px solid transparent");

    if (UserFIO.length <= 3) {
        $("#EnableFioError").addClass("warning-error");
        $("#FioError").append("The name is too short");
        return;
    }

    if (WalletAddr.length <= 3) {
        $("#WalletAddr").css("border", "1px solid rgb(193, 30, 15)");
        return;
    }

    formData = {
        data: "addNewContact" + "," + session + "," + userId + "," + UserFIO + "," + About + "," + CoinId + "," + WalletAddr,
    };
    $.post("http://" + HOST_IP + ":21001", formData, function (data) {
        displayContacts(data);
        $("#ContactCreated").css("display", "block");
        setTimeout(() => {
            window.location.reload();
        }, 3000);
    });

}

function walletsInside(wallet, address) {
    var filter = 0;

    var coinId;
    coinId = CoinID[wallet];

    $("#w_selected_names").empty();
    $("#WalletType").empty();

    $("#WalletType").append("Your " + CoinById[coinId].Name + " adress: ");
    $("#w_selected_names").append(CoinById[coinId].Name);
    $("#w_selected_img2").attr("src", CoinById[coinId].ImgUrl);
    $("#wallet_addres").val(address);

    param = {
        data: 'balanceW,' + session + ',' + userId + ',' + coinId + ',,,'
    };
    $.post("http://" + HOST_IP + ":10000", param, function (data) {
        [balance, balanceUSD] = data.split(" ");
        balance = convertToDot(balance, CoinById[coinId].Scale, coinId);
        balanceUSD = parseFloat(balanceUSD / CoinById[coinId].priceDelimiter).toFixed(2);

        var stringdata1 = balance + "<span class=\"wallet-inside__unit\" style='top: -15px;'>" + CoinById[coinId].ShortName + "</span>";
        var stringdata2 = balanceUSD + " <span class=\"wallet-inside__unit\">USD</span>";

        $("#wallet_balance").empty();
        $("#wallet_balance").append(stringdata1);
        $("#Wallet_balance_inUSD").empty();
        $("#Wallet_balance_inUSD").append(stringdata2);

    });

    DisplayWalletHistory(address, coinId, "receive");
    displayQRCode(".modal__img", address);
}

function WalletActions() {
    setTimeout(() => {
        if ($(".my-wallet__title--input")) {
            //ReName Wallet
            // $(".my-wallet__title--input").on({
            //   focus: function () {
            //     if (!$(this).data("disabled")) this.blur();
            //   },
            //   dblclick: function () {
            //     $(this).data("disabled", true);
            //
            //     // тут я сделал так, чтоб найти карточку контакта и главному блоку добавить клас
            //     let thisContact = searchThisContact($(this));
            //     thisContact.classList.add("edit-name--on");
            //
            //     this.focus();
            //   },
            //   blur: function () {
            //     $(this).data("disabled", false);
            //     let thisContact = searchThisContact($(this));
            //     thisContact.classList.remove("edit-name--on");
            //     this.blur();
            //   },
            // });

            let flag = false;

            $(".my-wallet__edit-button").click(function (e) {
                let thisContact = e.target.parentNode.parentNode.parentNode;
                let content = e.target.parentNode.parentNode.children[1];
                let bottomBlock = e.target.parentNode.parentNode.childNodes[5].children;
                thisContact.classList.toggle("edit-ready--on");
                flag = false;
                if ($(".my-wallet__input--delete").length == 0) {
                    $(".btn-wallet-delete").removeAttr("disabled");
                }

                if (thisContact.classList.contains("edit-ready--on")) {
                    bottomBlock[0].innerText = "Options";
                } else {
                    bottomBlock[0].innerText = "Address";
                    bottomBlock[0].style.cssText = "";

                    let p = e.target.parentNode.parentNode.getElementsByClassName("my-wallet__message")[0];
                    let input = e.target.parentNode.parentNode.getElementsByClassName("my-wallet__input--delete")[0];
                    let head = e.target.parentNode.children;
                    if (p) {
                        p.remove();
                    }

                    head[1].style.display = "";
                    head[2].style.display = "";

                    content.style.display = "";
                    if (input) {
                        input.remove();
                    }
                    bottomBlock[2].children[0].style.display = "";
                    bottomBlock[2].children[1].style.cssText = "";
                }
                $("#btn-wallet-dublicate").css("display", "none");
            });

            $(".btn-wallet-delete").click(function (e) {
                e.preventDefault();
                let myWallet = this.form.parentNode;

                let p = document.createElement("p");
                let head = this.form.parentNode.children[0].children[0];
                let content = this.form.parentNode.children[0].children[1];
                let bottom = this.form.parentNode.children[0].children[2].children;

                if (myWallet.classList.contains("balance-is-there")) {
                    if ($(".my-wallet__input--delete").length == 0) {
                        p.classList.add("my-wallet__message");
                        p.innerHTML = "There is a balance in the wallet";

                        head.children[1].style.display = "none";
                        head.children[2].style.display = "none";
                        head.appendChild(p);

                        bottom[0].innerHTML = "Make a translation";
                        bottom[0].style.cssText =
                            "text-align: center; text-transform: uppercase";

                        bottom[2].children[0].style.display = "none";
                        bottom[2].children[1].style.display = "none";

                        bottom[2].children[2].style.cssText =
                            "display: flex; margin: 0 auto;";
                    }
                } else if (myWallet.classList.contains("no-balance")) {
                    if ($(".my-wallet__input--delete").length == 0) {
                        let input = document.createElement("input");
                        let walletName = myWallet.firstElementChild[0].value;

                        if (!flag) {
                            input.classList.add("my-wallet__input--delete");
                            input.placeholder = "Type DELETE";
                            input.id = "RedyToDel";
                            input.style.textTransform = "none";
                            p.classList.add("my-wallet__message");
                            p.innerHTML = `Delete wallet <span>${walletName}</span>`;

                            head.children[1].style.display = "none";
                            head.children[2].style.display = "none";

                            head.appendChild(p);

                            content.style.display = "none";

                            bottom[0].innerHTML = "Type text delete";
                            bottom[0].style.cssText =
                                "text-align: center; text-transform: uppercase";

                            bottom[2].children[0].style.display = "none";
                            bottom[2].children[1].style.margin = "0 auto";
                            bottom[0].parentNode.insertBefore(input, bottom[1]);

                            this.disabled = "true";
                            input.oninput = function () {
                                if (input.value == "DELETE") {
                                    $(".btn-wallet-delete").removeAttr("disabled");
                                }
                            };
                            // отключить действие
                            flag = true;
                        }
                    }
                }
            });

            // функция по поиску главного елемента контакта тоесть - class="contacts__item contact"
            function searchThisContact(hereThis) {
                return hereThis[0].form.parentNode;
            }
        }
    }, 500);
}

function CopyToClipboard(e) {
    var text = e.innerText.replace(/ /g, "");
    alert(text);
    let inp = document.createElement("input");
    document.body.appendChild(inp);
    inp.value = text;
    inp.select();
    inp.setSelectionRange(0, 99999);
    document.execCommand("copy", false);
    inp.remove();
}

function copyAdress() {
    let element = document.querySelectorAll(
        ".wallet-inside__copy-button"
    );
    const parentAddress = element[0].parentNode.children[0];
    copyToClipboard2(parentAddress, "");
}
function copyToClipboard2(copyElem, tooltip) {
    let copyText = copyElem;
    if (copyText.localName === "input") {
        check(copyText);
    } else {
        let inp = document.createElement("input");
        document.body.appendChild(inp);
        inp.value = copyText.textContent.replace(/ /g, "");
        inp.remove();
    }

    function check(whatElem) {
        whatElem.select();
        whatElem.setSelectionRange(0, 99999);
        document.execCommand("copy", false);
    }

}

function loadPageSettings() {
    covertTo();
    setTimeout(function () {
        var CurrentPage = "";
        CurrentPage = window.location.pathname;
        CurrentPage = CurrentPage.replace("/", "");
        if (CurrentPage != "" && CurrentPage != "sign-up.html") {
            replaceOnPage();
        }
        if (CurrentPage == "wallets.html") {
            thisIsDemo();
            WalletsTemplate = $('#all-wallets-template').html();
            TepmpalteDIV = "#display_all_wallets";
            displayAllWallets("wallets", "1");
            WalletActions();
        }
        if (CurrentPage == "transactions.html") {
            thisIsDemo();
            $("#receive-noTr").css("display", "block");
            $("#send-noTr").css("display", "block");
            WalletsTemplate = $('#transactions_all_w_template').html();
            TepmpalteDIV = "#transactions_recive_w_add";
            displayAllWallets("transactions", "1");
            var sendPage = findGetParameter("send");
            var recive = findGetParameter("recive");

            if (sendPage == "true") {
                $("#recive").removeClass("is-active");
                $("#recive_link").removeClass("is-active");
                $("#send").addClass("is-active");
                $("#send_link").addClass("is-active");
            }
            if (recive == "true") {
                $("#send").removeClass("is-active");
                $("#send_link").removeClass("is-active");
                $("#recive").addClass("is-active");
            }
        }
        if (CurrentPage == "history.html") {
            thisIsDemo();
            WalletsTemplate = $('#all-history-template').html();
            TepmpalteDIV = "#all_history_template";
            displayAllHistory("history", "1");
        }
        if (CurrentPage == "wallet-inside.html") {
            thisIsDemo();
            $("#receive-noTr").css("display", "block");
            var wallet = findGetParameter("wallet");
            var address = findGetParameter("addr");
            walletsInside(wallet, address);
        }
        if (CurrentPage == "contacts.html") {
            thisIsDemo();
            displayContacts("");
            saveImg();
        }
        if (CurrentPage == "dashboard.html") {
            thisIsDemo();
            getOverviewData();
            WalletsTemplate = $('#r_transactions-card-template').html();
            TepmpalteDIV = "#r_transactions-content-dashboard";
            displayAllHistory("dashboard", "1");
            displayAllWallets("main", "1");
            loadMarkets();
        }
    }, 3);
}

function addmodalsAll() {
    const overlayModal = document.querySelector(".modal-overlay");
    const modals = document.querySelectorAll(".modal");
    const triggerModal = document.querySelectorAll(".modal-trigger");
    const closersModal = document.querySelectorAll("[data-close]");
    const body = document.querySelector("body");

    // let modalsLen = modals.length;
    let triggerModalLen = triggerModal.length;

    if (triggerModal.length !== 0) {
        for (let i = 0; i < triggerModalLen; i++) {
            if (triggerModal[i] != null) {
                triggerModal[i].addEventListener("click", getId, false);
                overlayModal.addEventListener("click", closeOverlay, false);
            }
        }
        for (let i = 0; i < closersModal.length; i++) {
            closersModal[i].addEventListener("click", close, false);
        }

        function getId(event) {
            event.preventDefault();
            let self = this;
            // get the value of the data-modal attribute from the button
            let modalId = self.dataset.modal;
            let len = modalId.length;
            // remove the '#' from the string
            let modalIdTrimmed = modalId.substring(1, len);
            // select the modal we want to activate
            let modal = document.getElementById(modalIdTrimmed);

            overlayModal.classList.add("active");
            modal.classList.add("active");
            body.classList.add("is-hidden");

            if ($("#created_wallet_address") != null) {
                $("#created_wallet_address").empty();
                $("#created_wallet_address").val("...");
            }
        }

        function close(event) {
            event.preventDefault();
            let self = this;
            let modalActive = document.querySelector(".modal.active");

            if (modalActive) {
                modalActive.classList.remove("active");
                overlayModal.classList.remove("active");
                body.classList.remove("is-hidden");
            }
        }

        function closeOverlay(event) {
            let overlayModalActive = document.querySelector(".modal-overlay.active");
            let modalActive = document.querySelector(".modal.active");
            if (overlayModalActive) {
                if (overlayModal === event.target) {
                    modalActive.classList.remove("active");
                    overlayModalActive.classList.remove("active");
                    body.classList.remove("is-hidden");
                }
            }
        }
    }
}

window.onload = function ($) {
    init_data();
};

function subscribeNews() {
    var email = $("#subscribeNews").val();
    var is_ok = validateEmail(email);
    $("#modal__desc").empty();
    if (is_ok) {
        param = {
            data: "Subscribe" + "," + session + "," + userId + "," + "," + "," + ",",
            email: email
        };
        $.post("http://" + HOST_IP + ":20001", param, function (data) {
            if (data == "1") {
                $("#almost-done2").addClass("active");
                $("#modal_overlay").addClass("active");
            } else {
                $("#attention3").addClass("active");
                $("#modal_overlay").addClass("active");
                $("#modal__desc").append("Something Wrong...");
            }
        });
    } else {
        $("#attention3").addClass("active");
        $("#modal_overlay").addClass("active");
        $("#modal__desc").append("Wrong E-mail.");
    }
}

function submitRequest() {
    var email = $("#RequestEmail").val().replace(/ /g, "");
    var is_ok = validateEmail(email);
    var issue = $("#DiscribeProblem").val();
    var wallet = $("#coin-wallet-select").val();
    // var wallet_address = $("#wallet-address").val();
    var description = $("#description").val();
    var descriptioncmp = description.replace(/ /g, "");
    $("#modal__desc").empty();
    $("#modal__desc2").empty();
    if (is_ok) {
        if (issue.length > 3 && wallet.length > 3 && descriptioncmp.length > 15) {
            param = {
                data: "SupportRequest" + "," + session + "," + userId + "," + "," + "," + ",",
                email: email,
                issue: issue,
                wallet: wallet,
                // wallet_address: wallet_address,
                wallet_address: "-",
                description: description,
            };
            $.post("http://" + HOST_IP + ":25001", param, function (data) {
                var response = "Request № "+ data + ". Thank you for contacting our Support team. We will answer you as soon as we process your request.";
                $("#modal__desc2").append(response);
                $("#almost-done2").addClass("active");
                $("#modal_overlay").addClass("active");
            });
        } else {
            if (issue.length < 3) {
                $("#attention3").addClass("active");
                $("#modal_overlay").addClass("active");
                $("#modal__desc").append("issue description is too short</br>");
            }
            if (wallet.length < 3) {
                $("#attention3").addClass("active");
                $("#modal_overlay").addClass("active");
                $("#modal__desc").append("Wallet type is too short</br>");
            }
            // if (wallet_address.length < 30) {
            //     $("#attention3").addClass("active");
            //     $("#modal_overlay").addClass("active");
            //     $("#modal__desc").append("Wallet address is too short</br>");
            // }
            if (descriptioncmp.length < 15) {
                $("#attention3").addClass("active");
                $("#modal_overlay").addClass("active");
                $("#modal__desc").append("Description is too short</br>");
            }
        }

    } else {
        $("#attention3").addClass("active");
        $("#modal_overlay").addClass("active");
        $("#modal__desc").append("Wrong E-mail.");
    }
}

function thisIsDemo() {
    var ex2walletsDemo = getCookie("ex2walletsDemo");
    if (ex2walletsDemo == null) {
        $("#thisIsDemo").addClass("active");
        $("#modal_overlay").addClass("active");
        setCookie("ex2walletsDemo", "ex2walletsDemo", 9999999999); //session for 1 day
    }
}

function init_data() {
    loadSettings();
    check_cookie();
    loadPageSettings();
    addmodalsAll();
}

function searchSubject() {}

function changePassword() {}

function getHistoryofUserSessions() {}

//0,0001
//Atrium Noctis - Datura Noir.