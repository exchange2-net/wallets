"use strict";

document.addEventListener("DOMContentLoaded", () => {
  // handlebars Dashboard
  {
    // overview-template
    {
      // let overview = jQuery("#overview").html();
      // let overviewTemplate = jQuery("#overview-template").html();
      //
      // let dataObjectOverview = [
      //   {
      //     mainCurrencyPercent: "+10"
      //   }
      // ];
      //
      // if (overviewTemplate) {
      //   for (let ii = 0; ii < dataObjectOverview.length; ii++) {
      //     addDataOverview(
      //       dataObjectOverview[ii],
      //       overviewTemplate,
      //       "overview",
      //       ii
      //     );
      //   }
      // }
      //
      // function addDataOverview(
      //   dataObjectOverview,
      //   handlebars_template,
      //   template_id,
      //   ii
      // ) {
      //   // let str = handlebars_template;
      //   let overview = handlebars_template.replace(/{mainCurrencyPercent}/g, dataObjectOverview.mainCurrencyPercent
      //     );
      //   document.getElementById(template_id).innerHTML = overview;
      // }
    }

    // ballance-template
    {
      let dataObjectBallance = [
        [
          {
            percentage: 70,
            ballanceName: "Bitcoin"
          },
          {
            percentage: 15,
            ballanceName: "Ethereum"
          },
          {
            percentage: 25,
            ballanceName: "Litecoin"
          },
          {
            percentage: 35,
            ballanceName: "Dash"
          }
        ]
      ];

      for (let ii = 0; ii < dataObjectBallance[0].length; ii++) {
        addDataBallance(dataObjectBallance[ii], "ballance-list", ii);
      }

      function addDataBallance(dataObjectBallance, template_id, ii) {
        for (let key in dataObjectBallance) {
          let li = document.createElement("li");
          li.classList.add("ballance__item");

          let ballanceTemplate = jQuery("#ballance-template").html();

          if (ballanceTemplate) {
            ballanceTemplate = ballanceTemplate
              .replace(/{percentage}/g, dataObjectBallance[key].percentage)
              .replace(/{ballanceName}/g, dataObjectBallance[key].ballanceName);
            li.innerHTML = ballanceTemplate;
            document.getElementById(template_id).appendChild(li);
          }
        }
      }
    }

    // transaction
    {
      let dataObjectTransaction = [
        [
          {
            status: "is-send",
            date: "16:23, 12 dec 2018",
            amount_in_wallet: 0.00,
            amount_in_wallet_currency: "btc",
            amount_equivalent: 3345.78,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-received",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.000,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-pending",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.109,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-canceled",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.109,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-canceled",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.109,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-pending",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.109,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-pending",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.109,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-send",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.109,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          }
        ]
      ];

      for (let ii = 0; ii < dataObjectTransaction[0].length; ii++) {
        addDataTransaction(
          dataObjectTransaction[ii],
          "transactions-content-dashboard",
          ii
        );
      }

      function addDataTransaction(dataObjectTransaction, template_id, ii) {
        for (let key in dataObjectTransaction) {
          let div = document.createElement("div");
          div.classList.add("transactions__card");

          let transactionsTemplate = jQuery(
            "#transactions-card-template"
          ).html();

          if (transactionsTemplate) {
            transactionsTemplate = transactionsTemplate
              .replace(/{date}/g, dataObjectTransaction[key].date)
              .replace(
                /{amount_in_wallet}/g,
                dataObjectTransaction[key].amount_in_wallet
              )
              .replace(
                /{amount_in_wallet_currency}/g,
                dataObjectTransaction[key].amount_in_wallet_currency
              )
              .replace(
                /{amount_equivalent}/g,
                dataObjectTransaction[key].amount_equivalent
              )
              .replace(
                /{amount_equivalent_currency}/g,
                dataObjectTransaction[key].amount_equivalent_currency
              )
              .replace(
                /{transactions_id}/g,
                dataObjectTransaction[key].transactions_id
              )
              .replace(
                /{transactions_from}/g,
                dataObjectTransaction[key].transactions_from
              )
              .replace(
                /{transactions_to}/g,
                dataObjectTransaction[key].transactions_to
              );
            div.classList.add(dataObjectTransaction[key].status);
            div.innerHTML = transactionsTemplate;

            document.getElementById(template_id).appendChild(div);
          }
        }
      }
    }
  }

  // handlebars Wallets
  {
    // wallet-inside balance
    {
      let walletBalanceTemplate = jQuery(
        "#wallet-inside-balance-template"
      ).html();

      let dataObjectWalletBalance = [
        {
          currency_balance: 0.221746,
          currency_balance_unit: "btc",
          equivalent: 0.3456789,
          equivalent_unit: "usd"
        }
      ];

      if (walletBalanceTemplate) {
        for (let ii = 0; ii < dataObjectWalletBalance.length; ii++) {
          addDataWalletBalance(
            dataObjectWalletBalance[ii],
            walletBalanceTemplate,
            "wallet-inside-balance",
            ii
          );
        }
      }

      function addDataWalletBalance(
        dataObjectWalletBalance,
        handlebars_template,
        template_id,
        ii
      ) {
        // let str = handlebars_template;
        let walletBalance = handlebars_template
          .replace(
            /{currency_balance}/g,
            dataObjectWalletBalance.currency_balance
          )
          .replace(
            /{currency_balance_unit}/g,
            dataObjectWalletBalance.currency_balance_unit
          )
          .replace(/{equivalent}/g, dataObjectWalletBalance.equivalent)
          .replace(
            /{equivalent_unit}/g,
            dataObjectWalletBalance.equivalent_unit
          );
        document.getElementById(template_id).innerHTML = walletBalance;
      }
    }
    // transaction
    {
      let dataObjectTransactionWallet = [
        [
          {
            status: "is-send",
            date: "16:23, 12 dec 2018",
            amount_in_wallet: 0.000,
            amount_in_wallet_currency: "btc",
            amount_equivalent: 3345.78,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-received",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.009,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-pending",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.109,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-canceled",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.109,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-canceled",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.109,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-pending",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.109,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-pending",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.109,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },
          {
            status: "is-send",
            date: "16:23, 12 dec 2019",
            amount_in_wallet: 0.109,
            amount_in_wallet_currency: "eth",
            amount_equivalent: 3345.88,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          }
        ]
      ];

      for (let ii = 0; ii < dataObjectTransactionWallet[0].length; ii++) {
        addDataTransaction(
          dataObjectTransactionWallet[ii],
          "transactions-content-wallet-inside",
          ii
        );
      }

      function addDataTransaction(
        dataObjectTransactionWallet,
        template_id,
        ii
      ) {
        for (let key in dataObjectTransactionWallet) {
          let div = document.createElement("div");
          div.classList.add("transactions__card");

          let transactionsWalletTemplate = jQuery(
            "#transactions-card-wallet-inside-template"
          ).html();

          if (transactionsWalletTemplate) {
            transactionsWalletTemplate = transactionsWalletTemplate
              .replace(/{date}/g, dataObjectTransactionWallet[key].date)
              .replace(
                /{amount_in_wallet}/g,
                dataObjectTransactionWallet[key].amount_in_wallet
              )
              .replace(
                /{amount_in_wallet_currency}/g,
                dataObjectTransactionWallet[key].amount_in_wallet_currency
              )
              .replace(
                /{amount_equivalent}/g,
                dataObjectTransactionWallet[key].amount_equivalent
              )
              .replace(
                /{amount_equivalent_currency}/g,
                dataObjectTransactionWallet[key].amount_equivalent_currency
              )
              .replace(
                /{transactions_id}/g,
                dataObjectTransactionWallet[key].transactions_id
              )
              .replace(
                /{transactions_from}/g,
                dataObjectTransactionWallet[key].transactions_from
              )
              .replace(
                /{transactions_to}/g,
                dataObjectTransactionWallet[key].transactions_to
              );
            div.classList.add(dataObjectTransactionWallet[key].status);
            div.innerHTML = transactionsWalletTemplate;

            document.getElementById(template_id).appendChild(div);
          }
        }
      }
    }
  }

  // handlebars Transaction
  {
    // transaction
    {
      let dataObjectTransactionPage = [
        [
          {
            status: "is-send",
            date: "16:23, 12 dec 2018",
            amount_in_wallet: 0.000,
            amount_in_wallet_currency: "btc",
            amount_equivalent: 3345.78,
            amount_equivalent_currency: "usd",
            transactions_id: "234567890-098765432",
            transactions_from: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB",
            transactions_to: "1PRj85hu9RXPZTzxtko9stfs6nRo1vyrQB"
          },

        ]
      ];

      for (let ii = 0; ii < dataObjectTransactionPage[0].length; ii++) {
        addDataTransactionPage(
          dataObjectTransactionPage[ii],
          "transactions-template-for-fiat",
          ii
        );
      }

      for (let ii = 0; ii < dataObjectTransactionPage[0].length; ii++) {
        addDataTransactionPage(
          dataObjectTransactionPage[ii],
          "transactions-template-for-send",
          ii
        );
      }

      for (let ii = 0; ii < dataObjectTransactionPage[0].length; ii++) {
        addDataTransactionPage(
          dataObjectTransactionPage[ii],
          "transactions-template-for-recive",
          ii
        );
      }

      function addDataTransactionPage(dataObjectTransactionPage, template_id, ii) {
        for (let key in dataObjectTransactionPage) {
          let div = document.createElement("div");
          div.classList.add("transactions__card");

          let transactionsTemplate = jQuery(
            "#transactions-page-card-template"
          ).html();

          if (transactionsTemplate) {
            transactionsTemplate = transactionsTemplate
              .replace(/{date}/g, dataObjectTransactionPage[key].date)
              .replace(
                /{amount_in_wallet}/g,
                dataObjectTransactionPage[key].amount_in_wallet
              )
              .replace(
                /{amount_in_wallet_currency}/g,
                dataObjectTransactionPage[key].amount_in_wallet_currency
              )
              .replace(
                /{amount_equivalent}/g,
                dataObjectTransactionPage[key].amount_equivalent
              )
              .replace(
                /{amount_equivalent_currency}/g,
                dataObjectTransactionPage[key].amount_equivalent_currency
              )
              .replace(
                /{transactions_id}/g,
                dataObjectTransactionPage[key].transactions_id
              )
              .replace(
                /{transactions_from}/g,
                dataObjectTransactionPage[key].transactions_from
              )
              .replace(
                /{transactions_to}/g,
                dataObjectTransactionPage[key].transactions_to
              );
            div.classList.add(dataObjectTransactionPage[key].status);
            div.innerHTML = transactionsTemplate;

            document.getElementById(template_id).appendChild(div);
          }
        }
      }
    }
  }
});
