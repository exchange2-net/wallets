"use strict";

document.addEventListener("DOMContentLoaded", () => {
  window.addEventListener("resize", () => {
    // for mobile device - really height 100%
    let vh = window.innerHeight * 0.01;
    document.documentElement.style.setProperty("--vh", `${vh}px`);
  });

  // mobile-menu
  {
    let burgerMenu = document.querySelector(".burger-menu");
    let body = document.querySelector("body");
    let header = document.querySelector(".header");

    if (burgerMenu) {
      burgerMenu.addEventListener("click", function () {
        header.classList.toggle("is-fixed");
        body.classList.toggle("is-hidden");
      });
    }
  }


  // DATA SCRIPT
  {
    let dateFrom = new FooPicker({
      id: "dateFrom",
    });
    let dateTo = new FooPicker({
      id: "dateTo",
    });
  }

  // MODAL SCRIPT
  // {
  //   setTimeout(() => {
  //     const overlayModal = document.querySelector(".modal-overlay");
  //     const modals = document.querySelectorAll(".modal");
  //     const triggerModal = document.querySelectorAll(".modal-trigger");
  //     const closersModal = document.querySelectorAll("[data-close]");
  //     const body = document.querySelector("body");
  //
  //     // let modalsLen = modals.length;
  //     let triggerModalLen = triggerModal.length;
  //
  //     if (triggerModal.length !== 0) {
  //       for (let i = 0; i < triggerModalLen; i++) {
  //         if (triggerModal[i] != null) {
  //           triggerModal[i].addEventListener("click", getId, false);
  //           overlayModal.addEventListener("click", closeOverlay, false);
  //         }
  //       }
  //       for (let i = 0; i < closersModal.length; i++) {
  //         closersModal[i].addEventListener("click", close, false);
  //       }
  //
  //       function getId(event) {
  //         event.preventDefault();
  //         let self = this;
  //         // get the value of the data-modal attribute from the button
  //         let modalId = self.dataset.modal;
  //         let len = modalId.length;
  //         // remove the '#' from the string
  //         let modalIdTrimmed = modalId.substring(1, len);
  //         // select the modal we want to activate
  //         let modal = document.getElementById(modalIdTrimmed);
  //
  //         overlayModal.classList.add("active");
  //         modal.classList.add("active");
  //         body.classList.add("is-hidden");
  //
  //         if ($("#created_wallet_address") != null) {
  //           $("#created_wallet_address").empty();
  //           $("#created_wallet_address").val("...");
  //         }
  //       }
  //
  //       function close(event) {
  //         event.preventDefault();
  //         let self = this;
  //         let modalActive = document.querySelector(".modal.active");
  //
  //         if (modalActive) {
  //           modalActive.classList.remove("active");
  //           overlayModal.classList.remove("active");
  //           body.classList.remove("is-hidden");
  //         }
  //       }
  //       function closeOverlay(event) {
  //         let overlayModalActive = document.querySelector(
  //           ".modal-overlay.active"
  //         );
  //         let modalActive = document.querySelector(".modal.active");
  //         if (overlayModalActive) {
  //           if (overlayModal === event.target) {
  //             modalActive.classList.remove("active");
  //             overlayModalActive.classList.remove("active");
  //             body.classList.remove("is-hidden");
  //           }
  //         }
  //       }
  //     }
  //   }, 500);
  // }

  // add-sending-amount
  {
    let buttonAddSendingAmount = document.querySelector("#add-sending-amount");
    let numAmount = 1;
    let mixerSenderCard = $(".mixer-sender__card").clone();

    if (buttonAddSendingAmount) {
      const mixerConent = new SimpleBar(
        document.querySelector(".mixer-sender__blocks-amount")
      );
      let thisContent = $(".mixer-sender__blocks-amount .simplebar-content");
      let container = document.querySelector(
        ".mixer-sender__blocks-amount .simplebar-content-wrapper"
      );

      buttonAddSendingAmount.addEventListener("click", function () {
        ++numAmount;
        let newNumAmount = `Amount #${numAmount}`;
        mixerSenderCard
          .find(".mixer-sender__text.sender-amount")
          .html(newNumAmount);
        thisContent.append(mixerSenderCard.clone());
        let thisHeight = thisContent.prop("scrollHeight");
        container.scrollTo({ top: thisHeight, behavior: "smooth" });

        // mixerConent.getScrollElement().scrollTop = ;
      });
    }
  }

  // CUSTOM SELECT "SORTING"
  {
    let customSelect = document.querySelectorAll(".sorting");
    let customOption = document.querySelectorAll(".sorting__item");
    let letiantSelect = document.querySelector(".sorting__text");

    if (customSelect.length !== 0) {
      for (const dropdown of customSelect) {
        dropdown.addEventListener("click", function () {
          this.classList.toggle("sorting__visible");
          this.querySelector(".sorting__dropdown").classList.toggle(
            "dropdown-hidden"
          );
        });
      }

      for (const option of customOption) {
        option.addEventListener("click", function () {
          if (!this.classList.contains("is-selected")) {
            this.parentNode
              .querySelector(".sorting__item.is-selected")
              .classList.remove("is-selected");
            this.classList.add("is-selected");
            this.closest(".sorting").querySelector(
              ".sorting__selected"
            ).innerHTML = this.firstElementChild.innerHTML;
          }
        });
      }

      window.addEventListener("click", function (e) {
        for (const select of customSelect) {
          if (!select.contains(e.target)) {
            // let thisDropdown = document.querySelector(".sorting__dropdown");
            select.classList.remove("sorting__visible");
            select.children[1].classList.add("dropdown-hidden");
          }
        }
      });
    }
  }

  // "SWITCH" in dashboard chart

  {
    let switchSelect = document.querySelectorAll(".switch");
    let switchOption = document.querySelectorAll(".switch__item");

    if (switchSelect.length !== 0) {
      for (const sw_option of switchOption) {
        sw_option.addEventListener("click", function () {
          if (!this.classList.contains("is-selected")) {
            this.parentNode
              .querySelector(".switch__item.is-selected")
              .classList.remove("is-selected");
            this.classList.add("is-selected");
          }
        });
      }
    }
  }

  // VALIDATION INPUT
  {
    // PAY-CARD
    {
      let transactions_page = document.querySelector(".transactions-page");

      if (transactions_page) {
        let input_card_number = document.querySelectorAll(
          ".bank-card__input-number"
        );
        let input_card_month = document.querySelector(
          ".bank-card__input-month"
        );
        let input_card_year = document.querySelector(".bank-card__input-year");
        let input_card_cvv = document.querySelector(".bank-card__input-cvv");

        input_card_month.addEventListener("keydown", function (e) {
          let value = this.value.replace(/\s+/g, "");
          let isBackspace = e.key === "Backspace";

          if (
            (e.key.length === 1 && /^[^\d\s]+$/.test(e.key)) ||
            (!isBackspace && value.length === 2)
          ) {
            e.preventDefault();
            return false;
          }

          this.value = value
            .split("")
            .reverse()
            .join("")
            .replace(/\B(?=(\d{2})+(?!\d))/g)
            .split("")
            .reverse()
            .join("")
            .trim();
        });
        input_card_year.addEventListener("keydown", function (e) {
          let value = this.value.replace(/\s+/g, "");
          let isBackspace = e.key === "Backspace";

          if (
            (e.key.length === 1 && /^[^\d\s]+$/.test(e.key)) ||
            (!isBackspace && value.length === 2)
          ) {
            e.preventDefault();
            return false;
          }

          this.value = value
            .split("")
            .reverse()
            .join("")
            .replace(/\B(?=(\d{2})+(?!\d))/g)
            .split("")
            .reverse()
            .join("")
            .trim();
        });
        input_card_cvv.addEventListener("keydown", function (e) {
          let value = this.value.replace(/\s+/g, "");
          let isBackspace = e.key === "Backspace";

          if (
            (e.key.length === 1 && /^[^\d\s]+$/.test(e.key)) ||
            (!isBackspace && value.length === 3)
          ) {
            e.preventDefault();
            return false;
          }

          this.value = value
            .split("")
            .reverse()
            .join("")
            .replace(/\B(?=(\d{3})+(?!\d))/g)
            .split("")
            .reverse()
            .join("")
            .trim();
        });

        for (let i = 0; i < input_card_number.length; i++) {
          const element = input_card_number[i];
          element.addEventListener("keydown", function (e) {
            let value = this.value.replace(/\s+/g, "");
            let isBackspace = e.key === "Backspace";

            if (
              (e.key.length === 1 && /^[^\d\s]+$/.test(e.key)) ||
              (!isBackspace && value.length === 4)
            ) {
              e.preventDefault();
              return false;
            }

            this.value = value
              .split("")
              .reverse()
              .join("")
              .replace(/\B(?=(\d{4})+(?!\d))/g)
              .split("")
              .reverse()
              .join("")
              .trim();
          });
        }
      }
    }
  }

  // TABS SCRIPT
  {
    let i, tabcontent, tablinks;

    tabcontent = document.getElementsByClassName("js-tabcontent");

    tablinks = document.getElementsByClassName("js-tab-links");

    for (i = 0; i < tablinks.length; i++) {
      tablinks[i].addEventListener("click", getId, false);
    }

    function getId(event) {
      event.preventDefault();
      let self = this;

      removeClass();

      for (i = 0; i < tabcontent.length; i++) {
        tabcontent[i].classList.remove("is-active");
        self.className.replace(" is-active", "");
      }
      // get the value of the data-modal attribute from the button
      let modalId = self.dataset.tab;
      let len = modalId.length;
      // remove the '#' from the string
      let modalIdTrimmed = modalId.substring(1, len);

      // select the modal we want to activate

      let modal = document.getElementById(modalIdTrimmed);

      self.className += " is-active";
      modal.classList.add("is-active");
      if ($("#created_wallet_address") != null) {
        $("#created_wallet_address").empty();
      }
    }

    function removeClass() {
      let current = document.querySelectorAll(".js-tab-links.is-active");
      if (current.length > 0) {
        current[0].className = current[0].className.replace(" is-active", "");
      }
    }
  }

  // SESSION TABLE MOBILE
  {
    let table_row = document.querySelectorAll(".sessions__table-row");
    if (table_row) {
      for (let i = 0; i < table_row.length; i++) {
        const element = table_row[i];

        element.addEventListener("click", function () {
          console.log(this);
          this.classList.toggle("is-opened");
        });
      }
    }
  }

  // copyToClipboard
  {
    setTimeout(() => {
      function copyToClipboard(copyElem, tooltip) {
        let copyText = copyElem;

        if (copyText.localName === "input") {
          check(copyText);
          // console.log("input");
        } else {
          let inp = document.createElement("input");

          document.body.appendChild(inp);
          inp.value = copyText.textContent.replace(/ /g, "");

          check(inp);

          inp.remove();
        }

        function check(whatElem) {
          whatElem.select();
          whatElem.setSelectionRange(0, 99999);
          document.execCommand("copy", false);
        }

        $(tooltip).fadeIn(300);
        setTimeout(function () {
          $(tooltip).fadeOut(300);
        }, 1000);
      }

      let btnCopyInputTextTrans = document.querySelectorAll(
        ".your-adress__copy-button"
      );
      let btnCopyInputText = document.querySelectorAll(
        ".create-wallet__copy-button"
      );
      let btnCopyWalletAddress = document.querySelectorAll(
        ".my-wallet__copy-button"
      );
      let btnCopyWalletAddress2 = document.querySelectorAll(
          ".wallet-inside__copy-button"
      );

      if (btnCopyInputText) {
        copyAdress(btnCopyInputText, 2);
      }
      if (btnCopyInputTextTrans) {
        copyAdress(btnCopyInputTextTrans, 2);
      }

      if (btnCopyWalletAddress) {
        copyAdress(btnCopyWalletAddress, 1);
      }
      if (btnCopyWalletAddress2) {
        copyAdress(btnCopyWalletAddress2, 1);
      }

      function copyAdress(element, id) {
        for (let i = 0; i < element.length; i++) {
          const thisWalletbtn = element[i];
          const parentAddress = element[i].parentNode.children[0];
          const thisTooltip = element[i].parentNode.children[id];

          thisWalletbtn.addEventListener("click", function (e) {
            e.preventDefault();
            copyToClipboard(parentAddress, thisTooltip);
          });
        }
      }
    }, 3000);
  }

  // contact name editable
  {
    setTimeout(() => {
      // if ($(".my-wallet__title--input")) {
      //   //ReName Wallet
      //   // $(".my-wallet__title--input").on({
      //   //   focus: function () {
      //   //     if (!$(this).data("disabled")) this.blur();
      //   //   },
      //   //   dblclick: function () {
      //   //     $(this).data("disabled", true);
      //   //
      //   //     // тут я сделал так, чтоб найти карточку контакта и главному блоку добавить клас
      //   //     let thisContact = searchThisContact($(this));
      //   //     thisContact.classList.add("edit-name--on");
      //   //
      //   //     this.focus();
      //   //   },
      //   //   blur: function () {
      //   //     $(this).data("disabled", false);
      //   //     let thisContact = searchThisContact($(this));
      //   //     thisContact.classList.remove("edit-name--on");
      //   //     this.blur();
      //   //   },
      //   // });
      //
      //   let flag = false;
      //
      //   $(".my-wallet__edit-button").click(function () {
      //     let thisContact = this.parentNode.parentNode.parentNode;
      //     let content = this.parentNode.parentNode.children[1];
      //     let bottomBlock = this.parentNode.parentNode.childNodes[5].children;
      //
      //     thisContact.classList.toggle("edit-ready--on");
      //
      //     flag = false;
      //     $(".btn-wallet-delete").removeAttr("disabled");
      //
      //     if (thisContact.classList.contains("edit-ready--on")) {
      //       bottomBlock[0].innerText = "Options";
      //     } else {
      //       bottomBlock[0].innerText = "Address";
      //       bottomBlock[0].style.cssText = "";
      //
      //       let p = $(".my-wallet__message");
      //       let input = $(".my-wallet__input--delete");
      //       let head = this.parentNode.children;
      //       p.remove();
      //
      //       head[1].style.display = "";
      //       head[2].style.display = "";
      //
      //       content.style.display = "";
      //
      //       input.remove();
      //       bottomBlock[2].children[0].style.display = "";
      //       bottomBlock[2].children[1].style.cssText = "";
      //       bottomBlock[2].children[2].style.display = "none";
      //     }
      //   });
      //
      //   $(".btn-wallet-delete").click(function (e) {
      //     e.preventDefault();
      //     let myWallet = this.form.parentNode;
      //
      //     let p = document.createElement("p");
      //     let head = this.form.parentNode.children[0].children[0];
      //     let content = this.form.parentNode.children[0].children[1];
      //     let bottom = this.form.parentNode.children[0].children[2].children;
      //
      //     if (myWallet.classList.contains("balance-is-there")) {
      //       p.classList.add("my-wallet__message");
      //       p.innerHTML = "There is a balance in the wallet";
      //
      //       head.children[1].style.display = "none";
      //       head.children[2].style.display = "none";
      //       head.appendChild(p);
      //
      //       bottom[0].innerHTML = "Make a translation";
      //       bottom[0].style.cssText =
      //         "text-align: center; text-transform: uppercase";
      //
      //       bottom[2].children[0].style.display = "none";
      //       bottom[2].children[1].style.display = "none";
      //
      //       bottom[2].children[2].style.cssText =
      //         "display: flex; margin: 0 auto;";
      //     } else if (myWallet.classList.contains("no-balance")) {
      //       let input = document.createElement("input");
      //       let walletName = myWallet.firstElementChild[0].value;
      //
      //       if (!flag) {
      //         input.classList.add("my-wallet__input--delete");
      //         input.placeholder = "Type DELETE";
      //
      //         p.classList.add("my-wallet__message");
      //         p.innerHTML = `Delete wallet <span>${walletName}</span>`;
      //
      //         head.children[1].style.display = "none";
      //         head.children[2].style.display = "none";
      //
      //         head.appendChild(p);
      //
      //         content.style.display = "none";
      //
      //         bottom[0].innerHTML = "Type text delete";
      //         bottom[0].style.cssText =
      //           "text-align: center; text-transform: uppercase";
      //
      //         bottom[2].children[0].style.display = "none";
      //         bottom[2].children[1].style.margin = "0 auto";
      //         bottom[0].parentNode.insertBefore(input, bottom[1]);
      //
      //         this.disabled = "true";
      //         input.oninput = function () {
      //           if (input.value == "delete") {
      //             $(".btn-wallet-delete").removeAttr("disabled");
      //           }
      //         };
      //         // отключить действие
      //         flag = true;
      //       } else {
      //         console.log("delete wallet");
      //       }
      //     } else {
      //       console.log("error");
      //     }
      //   });
      //
      //   // функция по поиску главного елемента контакта тоесть - class="contacts__item contact"
      //   function searchThisContact(hereThis) {
      //     return hereThis[0].form.parentNode;
      //   }
      // }

      if ($(".contact__name")) {
        $(".contact__name").on({
          focus: function () {
            if (!$(this).data("disabled")) this.blur();
          },
          dblclick: function () {
            $(this).data("disabled", true);

            // тут я сделал так, чтоб найти карточку контакта и главному блоку добавить клас
            let thisContact = searchThisContact($(this));
            thisContact.classList.add("edit--on");

            // тут я пробовал присвоить value в placeholder

            // let value = $(this).val()
            // $(this).attr('placeholder', value);
            // $(this).val("");

            this.focus();
          },

          // можно включить и тогда при клике на другую область стработает все то что на кнопку "enter"

          // blur: function () {
          //   $(this).data("disabled", false);
          //   $(this).removeClass("edit--on");
          //   this.blur();
          // },
        });

        // функция по отключения действия на "enter" в форме
        $(".contact__form-editable").keypress(function (event) {
          return event.keyCode != 13;
        });

        // изменения сотсояния input на клавишу "enter"
        $(".contact__name").on("keyup", function (e) {
          if (e.keyCode === 13) {
            $(this).data("disabled", false);

            let thisContact = searchThisContact($(this));

            thisContact.classList.remove("edit--on");
            // $(this).removeClass("edit--on");

            this.blur();
          }
        });

        // функция по поиску главного елемента контакта тоесть - class="contacts__item contact"
        function searchThisContact(hereThis) {
          return hereThis[0].form.parentNode.parentNode.parentNode;
        }
      }
    }, 500);
  }

  //* page - "Setting" tab "profile"
  {
    // * show block Password
    let btn_change_password = document.getElementById("change-password-btn");

    if (btn_change_password) {
      btn_change_password.addEventListener("click", function (e) {
        e.preventDefault();

        let span_desc = this.parentNode.parentNode.children[0].children[1];
        let block_pass = this.parentNode.parentNode.children[1];
        let span = this.parentNode.children[0];

        span_desc.classList.remove("is-hidden");
        span.classList.remove("is-hidden");
        block_pass.classList.add("is-visible");
      });
    }

    // * input type="password" or type="text"
    let eye = document.querySelectorAll(".eye");

    if (eye) {
      for (let i = 0; i < eye.length; i++) {
        const eye_elem = eye[i];

        eye_elem.addEventListener("click", togglePass);
      }
    }

    function togglePass() {
      let this_input = this.parentNode.children[1];

      this.classList.toggle("active");
      this_input.type == "password"
        ? (this_input.type = "text")
        : (this_input.type = "password");
    }
  }

  //* page - "Faq" accordion
  {
    // $(".accordion__answer:first").show();
    // $(".accordion__question:first").addClass("expanded");

    $(".accordion__question").on("click", function () {
      let content = $(this).next();

      $(".accordion__answer").not(content).slideUp(400);
      $(".accordion__question").not(this).removeClass("is-expanded");
      $(this).toggleClass("is-expanded");
      content.slideToggle(400);
    });
  }

  // * anchor function
  {
    $(".support-inside").on("click", 'a[href^="#"]', function (event) {
      event.preventDefault();

      $("html, body").animate(
        {
          scrollTop: $($.attr(this, "href")).offset().top,
        },
        1500
      );
    });

    const simpleBar_terms = document.querySelector(".terms__content");

    if (simpleBar_terms) {
      let n_bar = new SimpleBar(simpleBar_terms);

      n_bar.getScrollElement().addEventListener("scroll", function () {
        let $sections = $("h3");
        $sections.each(function (i, el) {
          let top = $(el).offset().top - 200;
          let bottom = top + $(el).height();
          let scroll = $(window).scrollTop();
          let id = $(el).attr("id");

          if (scroll > top && scroll < bottom) {
            $sections.css("color", "#3f4048");

            $("a.terms__anchor").removeClass("is-active");
            $('a[href="#' + id + '"]').addClass("is-active");

            // $("a.terms__anchor").css("color", "#3f4048");
            // $('a[href="#' + id + '"]').css("color", "#1428e5");

            el.style.color = "#1428e5";
          }
        });
      });

      $(".terms").on("click", 'a[href^="#"]', function (event) {
        event.preventDefault();

        let height_to = $($.attr(this, "href"))[0].offsetTop;

        $('a[href^="#"]').not(this).removeClass("is-active");

        $(this).addClass("is-active");
        for (let i = 0; i < $("h3").length; i++) {
          const h3 = $("h3")[i];
          if (h3.id === $($.attr(this, "href"))[0].id) {
            $($.attr(this, "href"))[0].style.color = "#1428e5";
          } else {
            h3.style.color = "#3f4048";
          }
        }

        n_bar.getScrollElement().scrollTo({
          top: height_to,
          behavior: "smooth",
        });
      });
    }
  }

  // * Markets
  {
    $(".markets__card").on("click", function (event) {
      event.preventDefault();
      $(".markets__card").not().removeClass("is-active");
      $(this).addClass("is-active");
    });
  }

  // * widt overview balance
  {
    setTimeout(() => {
      // текст по ширине родителя
      $("#OverviewBalanceUSD").each(function () {
        let length = $(this)
            .text()
            .replace(/^\s+|\s+$|\(|\)|8-/gm, "").length,
          size = ($(this).width() / length) * parseFloat($(this).data("ratio"));

        if (size >= 48) {
          $(this).css("font-size", "48px");
        } else if (size <= 16) {
          $(this).css("font-size", "16x");
        } else {
          $(this).css("font-size", size + "px");
        }
      });
    }, 3);
  }
});
