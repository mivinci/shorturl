window.addEventListener("DOMContentLoaded", () => {
  const form = document.getElementById("form");
  const input = document.getElementById("input");
  const select = document.querySelector(".input .select");
  const options = document.querySelectorAll(".input .option");
  const clock = document.querySelector(".input i");
  const result = document.getElementById("result");

  const TYPE_ERROR = 0, TYPE_TEXT = 1;
  const ttls = [-1, 3600, 86400, 604800, 2592000, 31536000];
  let selectShown = false;
  let selectedTTL = ttls[0];

  for (let i = 0; i < ttls.length; i++) {
    options[i].addEventListener("click", function () {
      selectedTTL = ttls[i];
      console.log(selectedTTL);
      setSelectedOption(i);
    })
  }

  function setSelectedOption(i) {
    options.forEach((option, index) => {
      if (index !== i && option.classList.contains("selected")) {
        option.classList.remove("selected");
      }
    });
    options[i].classList.add("selected");
  }

  clock.addEventListener("click", () => (selectShown ? eraseOptions() : drawOptions()))

  function eraseOptions() {
    selectShown = !selectShown;
    select.classList.remove("showup");
  }

  function drawOptions() {
    selectShown = !selectShown;
    select.classList.add("showup");
  }

  form.addEventListener("submit", function (evt) {
    evt.preventDefault();
    if (!input.value) return
    const uri = `?origin=${encodeURIComponent(input.value)}&ttl=${selectedTTL}`;
    throttle(500, fetchNewShortURL)(input.value, selectedTTL);
    eraseOptions();
  });

  function fetchNewShortURL(origin, ttl) {
    const uri = `?origin=${encodeURIComponent(origin)}&ttl=${ttl}`;
    fetch(uri, { method: "POST" })
      .then(res => {
        console.log(res.status)
        return res.text()
      })
      .then(res => (showResult(`<a href="${res}">${res}</a>`)))
      .catch(err => (showResult(err, TYPE_ERROR)));
    showResult("生成中，请稍等...");
  }

  function showResult(content, type = TYPE_TEXT) {
    if (type === TYPE_ERROR) {
      result.classList.add("error");
    }
    result.innerHTML = content;
  }

  function throttle(delay, func) {
    let timer = null;
    return function () {
      const args = arguments;
      if (!timer) {
        timer = setTimeout(() => {
          func.apply(this, args);
          timer = null;
        }, delay);
      }
    }
  }
})