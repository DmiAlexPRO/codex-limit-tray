# Codex Limit Tray

[English](README.md)

**Проверяйте лимиты Codex в трее Windows — на значке виден остаток в процентах.** Нажмите на него, чтобы посмотреть остаток и время сброса для пятичасового и недельного лимитов, или оставьте компактный виджет поверх рабочих окон.

**[Скачать для Windows · установщик x64](https://github.com/DmiAlexPRO/codex-limit-tray/releases/latest/download/CodexLimitTray-windows-x64-Setup.exe)** · [Все выпуски и переносные версии](https://github.com/DmiAlexPRO/codex-limit-tray/releases)

Скачивания за все выпуски на GitHub:

[![Скачивания: Установщик Windows](https://img.shields.io/github/downloads/DmiAlexPRO/codex-limit-tray/CodexLimitTray-windows-x64-Setup.exe?displayAssetName=false&label=%D0%A3%D1%81%D1%82%D0%B0%D0%BD%D0%BE%D0%B2%D1%89%D0%B8%D0%BA%20Windows)](https://github.com/DmiAlexPRO/codex-limit-tray/releases)
[![Скачивания: Переносная версия Windows](https://img.shields.io/github/downloads/DmiAlexPRO/codex-limit-tray/CodexLimitTray-windows-x64.zip?displayAssetName=false&label=%D0%9F%D0%B5%D1%80%D0%B5%D0%BD%D0%BE%D1%81%D0%BD%D0%B0%D1%8F%20%D0%B2%D0%B5%D1%80%D1%81%D0%B8%D1%8F%20Windows)](https://github.com/DmiAlexPRO/codex-limit-tray/releases)
[![Скачивания: macOS arm64](https://img.shields.io/github/downloads/DmiAlexPRO/codex-limit-tray/CodexLimitTray-macOS-arm64.zip?displayAssetName=false&label=macOS%20arm64)](https://github.com/DmiAlexPRO/codex-limit-tray/releases)
[![Скачивания: macOS x86_64](https://img.shields.io/github/downloads/DmiAlexPRO/codex-limit-tray/CodexLimitTray-macOS-x86_64.zip?displayAssetName=false&label=macOS%20x86_64)](https://github.com/DmiAlexPRO/codex-limit-tray/releases)

Бесплатный инструмент с открытым кодом под [лицензией MIT](LICENSE). Нужен установленный Codex со входом через поддерживаемую подписку ChatGPT. Проект независимый и не связан с OpenAI: он показывает **лимиты подписки Codex**, а не общие лимиты ChatGPT и не расходы на API.

![Четыре варианта значка в трее — цифры, бары, сегменты и кольца — на примере одних и тех же данных: осталось 82% пятичасового лимита и 57% недельного.](docs/images/tray-modes.png)

Интерфейс доступен на **27 языках**, в том числе на русском и английском. При установке или первом запуске приложение выбирает поддерживаемый язык системы. Если подходящего языка нет, используется английский. Чтобы сменить язык, нажмите шестерёнку → **Оформление → Язык** и перезапустите приложение.

![English](https://img.shields.io/badge/English-EN-2F7668)
![Simplified Chinese](https://img.shields.io/badge/%E7%AE%80%E4%BD%93%E4%B8%AD%E6%96%87-ZH-2F7668)
![Hindi](https://img.shields.io/badge/%E0%A4%B9%E0%A4%BF%E0%A4%A8%E0%A5%8D%E0%A4%A6%E0%A5%80-HI-2F7668)
![Spanish](https://img.shields.io/badge/Espa%C3%B1ol-ES-2F7668)
![Modern Standard Arabic](https://img.shields.io/badge/%D8%A7%D9%84%D8%B9%D8%B1%D8%A8%D9%8A%D8%A9%20%D8%A7%D9%84%D9%81%D8%B5%D8%AD%D9%89-AR-2F7668)
![French](https://img.shields.io/badge/Fran%C3%A7ais-FR-2F7668)
![Bengali](https://img.shields.io/badge/%E0%A6%AC%E0%A6%BE%E0%A6%82%E0%A6%B2%E0%A6%BE-BN-2F7668)
![Portuguese](https://img.shields.io/badge/Portugu%C3%AAs-PT-2F7668)
![Indonesian](https://img.shields.io/badge/Bahasa%20Indonesia-ID-2F7668)
![Urdu](https://img.shields.io/badge/%D8%A7%D8%B1%D8%AF%D9%88-UR-2F7668)
![Russian](https://img.shields.io/badge/%D0%A0%D1%83%D1%81%D1%81%D0%BA%D0%B8%D0%B9-RU-2F7668)
![German](https://img.shields.io/badge/Deutsch-DE-2F7668)
![Japanese](https://img.shields.io/badge/%E6%97%A5%E6%9C%AC%E8%AA%9E-JA-2F7668)
![Nigerian Pidgin](https://img.shields.io/badge/Naij%C3%A1-PCM-2F7668)
![Egyptian Arabic](https://img.shields.io/badge/%D8%A7%D9%84%D8%B9%D8%B1%D8%A8%D9%8A%D8%A9%20%D8%A7%D9%84%D9%85%D8%B5%D8%B1%D9%8A%D8%A9-ARZ-2F7668)
![Marathi](https://img.shields.io/badge/%E0%A4%AE%E0%A4%B0%E0%A4%BE%E0%A4%A0%E0%A5%80-MR-2F7668)
![Vietnamese](https://img.shields.io/badge/Ti%E1%BA%BFng%20Vi%E1%BB%87t-VI-2F7668)
![Swahili](https://img.shields.io/badge/Kiswahili-SW-2F7668)
![Hausa](https://img.shields.io/badge/Hausa-HA-2F7668)
![Belarusian](https://img.shields.io/badge/%D0%91%D0%B5%D0%BB%D0%B0%D1%80%D1%83%D1%81%D0%BA%D0%B0%D1%8F-BE-2F7668)
![Kazakh](https://img.shields.io/badge/%D2%9A%D0%B0%D0%B7%D0%B0%D2%9B%D1%88%D0%B0-KK-2F7668)
![Ukrainian](https://img.shields.io/badge/%D0%A3%D0%BA%D1%80%D0%B0%D1%97%D0%BD%D1%81%D1%8C%D0%BA%D0%B0-UK-2F7668)
![Polish](https://img.shields.io/badge/Polski-PL-2F7668)
![Italian](https://img.shields.io/badge/Italiano-IT-2F7668)
![Turkish](https://img.shields.io/badge/T%C3%BCrk%C3%A7e-TR-2F7668)
![Kyrgyz](https://img.shields.io/badge/%D0%9A%D1%8B%D1%80%D0%B3%D1%8B%D0%B7%D1%87%D0%B0-KY-2F7668)
![Armenian](https://img.shields.io/badge/%D5%80%D5%A1%D5%B5%D5%A5%D6%80%D5%A5%D5%B6-HY-2F7668)

Чтобы изменить значок в трее, нажмите шестерёнку → **Оформление → Значок** и выберите **Цифры**, **Бары**, **Сегменты** или **Кольца**.

![Пример: осталось 82% пятичасового лимита и 57% недельного. Под каждой шкалой указано время сброса.](docs/images/main-dark.png)

## Первый запуск в Windows

1. Установите Codex и войдите в аккаунт через Codex desktop или команду `codex login`. Отдельный аккаунт и API-ключ для утилиты не нужны.
2. Скачайте и запустите **CodexLimitTray-windows-x64-Setup.exe**. Программа установится для текущего пользователя без прав администратора. У установщика нет подписи издателя Windows (Authenticode), поэтому SmartScreen может показать предупреждение. Скачивайте его только из релизов этого репозитория.
3. Откройте **Codex Limit Tray**. Крестик прячет окно в трей; щелчок по значку возвращает его на экран. Если Codex не найден, нажмите шестерёнку → **Подключение** и введите полный путь к доверенному исполняемому файлу Codex.

Кнопка скачивания ведёт на последний **стабильный выпуск для Windows**, а не на предварительную сборку. Об обновлениях и ограничениях платформ — в [заметках к версии 3.0.1](docs/releases/v3.0.1.md).

## Какую версию выбрать?

| Вариант | Кому подойдёт | Обновления |
| --- | --- | --- |
| **Windows x64 · установщик** | Нужен ярлык в меню «Пуск» и обновление без ручной переустановки. | Стабильная версия по умолчанию проверяет при запуске только стабильные выпуски. Если вышла новая версия, проверяет подпись установщика, устанавливает обновление и возвращается в трей. Можно переключиться на ручную проверку. |
| **Windows x64 · переносной ZIP** | Удобнее распаковать EXE в свою папку и работать без установки. | Сам ничего не проверяет и не устанавливает. Можно заменить EXE вручную или явно выбрать установку обычной версии. Настройки сохранятся. |
| **macOS · экспериментально** | Для Apple Silicon выбирайте ZIP с пометкой `arm64`, для Intel — `x86_64`, если архив опубликован. | Распакуйте и установите или замените `.app` вручную. В приложении доступны проверка версии и описание релиза. Сборки без подписи и нотариального заверения Apple; проверка на реальных Mac ещё не завершена. |

В разделе **Настройки → Обновления** можно посмотреть, что изменилось, проверить наличие новой версии или отключить автоустановку. Если её выключить, при запуске не будет и проверки. Кнопка **Обновить лимиты** на главном экране обновляет только данные аккаунта. Перед установкой из приложения проверяется подпись Ed25519; она не удостоверяет подлинность первого скачанного EXE. [Подробнее об обновлениях](docs/UPDATES.md).

![Вкладка «Обновления»: номер установленной и доступной версии, кнопка установки и описание изменений. Номера версий приведены для примера.](docs/images/updates.png)

## Трей и виджет

- Наведите курсор на значок в трее, чтобы увидеть оба процента. Нажмите **Виджет**, чтобы оставить индикатор поверх окон; щелчок правой кнопкой по виджету скрывает его.
- В разделе **Оформление** можно выбрать тему, стиль значка и цвет низкого остатка, а в Windows — включить звуковой сигнал. Настройка **Запускать при входе в систему** открывает программу сразу в трее.
- Утилита показывает только те лимиты, которые вернул Codex. При неудачном запросе она покажет причину, а не устаревшие проценты.

![Компактный виджет с примером недельного остатка — 57%.](docs/images/widget.png)

## Какие данные передаются?

Утилита запускает локальный `codex app-server` и вызывает `account/rateLimits/read`. За авторизацию отвечает сам Codex; утилита не читает файлы токенов. Лимиты проверяются каждые 30 секунд; при ошибках пауза увеличивается до пяти минут. На каждый запрос пока запускается отдельный короткоживущий процесс.

**Аналитики нет**, история лимитов не сохраняется. При проверке обновлений программа обращается к GitHub, но не отправляет ему учётные данные Codex, сведения об аккаунте или значения лимитов. Настройки лежат в `CodexLimitTray/fyne-preview.json` внутри `%APPDATA%` в Windows или `~/Library/Application Support` в macOS. Путь в диагностике может содержать имя пользователя — скройте его перед публикацией. [Безопасность](SECURITY.md).

## Сборка и деморежим

Программа написана на Go и Fyne, без WebView. Нужны Go 1.26+ и компилятор с поддержкой CGO. В Windows потребуются MinGW/GCC с `windres` и PowerShell:

```powershell
.\build.ps1 -Preview
.\dist\CodexLimitTray.exe --demo
```

В macOS установите Xcode Command Line Tools и выполните на Mac нужной архитектуры:

```sh
sh scripts/build-macos.sh
open dist/CodexLimitTray.app --args --demo
```

`--demo` показывает пример без подключения к Codex и GitHub и без сохранения настроек. `--minimized` запускает программу в трее. Скрипты сборки выполняют `go test ./...` и `go vet ./...`. Как собрать установщик, подписать и проверить релиз — в [руководстве по выпуску](docs/RELEASING.md).

[История изменений](CHANGELOG.md) · [Как помочь проекту](CONTRIBUTING.md) · [Сторонние компоненты](THIRD_PARTY_NOTICES.txt)
