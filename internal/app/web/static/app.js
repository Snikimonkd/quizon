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
