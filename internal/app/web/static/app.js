// Обратный отсчёт до открытия регистрации.
// Сервер рендерит дату открытия, а за сутки до неё текст меняется на "через ЧЧ:ММ:СС".
// Когда время вышло, страница перезагружается, чтобы сервер показал открытую регистрацию.
(() => {
  const DAY = 24 * 60 * 60 * 1000;
  const RELOAD_GUARD_KEY = "quizon:countdown-reload";
  const pad = (n) => String(n).padStart(2, "0");

  const reload = () => {
    // если часы клиента спешат, не уходим в бесконечные перезагрузки
    const last = Number(sessionStorage.getItem(RELOAD_GUARD_KEY) || 0);
    if (Date.now() - last < 15000) {
      return;
    }
    sessionStorage.setItem(RELOAD_GUARD_KEY, String(Date.now()));
    location.reload();
  };

  const tick = () => {
    for (const el of document.querySelectorAll("time[data-countdown]")) {
      const left = Date.parse(el.dateTime) - Date.now();
      if (left <= 0) {
        el.removeAttribute("data-countdown");
        el.textContent = "уже сейчас";
        reload();
        continue;
      }
      if (left < DAY) {
        const s = Math.floor(left / 1000);
        el.textContent = `через ${pad(Math.floor(s / 3600))}:${pad(Math.floor(s / 60) % 60)}:${pad(s % 60)}`;
      }
    }
  };

  tick();
  setInterval(tick, 1000);
  // после htmx-свапа в разметке может появиться новый отсчёт
  document.addEventListener("htmx:after:swap", tick);
})();

// Подтверждение перед удалением (hx-confirm) своим окном вместо системного window.confirm.
// Вопрос берётся из hx-confirm, пояснение — из data-confirm-text того же элемента.
// Если окна на странице нет, htmx спросит как обычно, через window.confirm.
(() => {
  document.addEventListener("htmx:confirm", (event) => {
    const dialog = document.getElementById("confirm-dialog");
    if (!dialog) {
      return;
    }
    event.preventDefault();

    const { ctx, issueRequest, dropRequest } = event.detail;
    document.getElementById("confirm-title").textContent = ctx.confirm;
    document.getElementById("confirm-text").textContent = ctx.sourceElement.dataset.confirmText ?? "";
    // Esc и клик мимо окна закрывают его с пустым returnValue — это отмена
    dialog.returnValue = "";
    dialog.addEventListener("close", () => (dialog.returnValue === "confirm" ? issueRequest() : dropRequest()), {
      once: true,
    });
    dialog.showModal();
  });

  // клик по затемнению вокруг окна приходит в сам dialog, внутренности окна его не задевают
  document.addEventListener("click", (event) => {
    if (event.target instanceof HTMLDialogElement && event.target.id === "confirm-dialog") {
      event.target.close();
    }
  });
})();

// iOS Safari включает :active только на страницах с обработчиком касаний, без него нажатие на кнопку ничем не видно
document.addEventListener("touchstart", () => {}, { passive: true });
