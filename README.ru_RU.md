[English](/README.md) | [فارسی](/README.fa_IR.md) | [العربية](/README.ar_EG.md) |  [中文](/README.zh_CN.md) | [Español](/README.es_ES.md) | [Русский](/README.ru_RU.md)

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./media/3x-ui-dark.png">
    <img alt="3x-ui" src="./media/3x-ui-light.png">
  </picture>
</p>

[![Release](https://img.shields.io/github/v/release/mhsanaei/3x-ui.svg)](https://github.com/MHSanaei/3x-ui/releases)
[![Build](https://img.shields.io/github/actions/workflow/status/mhsanaei/3x-ui/release.yml.svg)](https://github.com/MHSanaei/3x-ui/actions)
[![GO Version](https://img.shields.io/github/go-mod/go-version/mhsanaei/3x-ui.svg)](#)
[![Downloads](https://img.shields.io/github/downloads/mhsanaei/3x-ui/total.svg)](https://github.com/MHSanaei/3x-ui/releases/latest)
[![License](https://img.shields.io/badge/license-GPL%20V3-blue.svg?longCache=true)](https://www.gnu.org/licenses/gpl-3.0.en.html)
[![Go Reference](https://pkg.go.dev/badge/github.com/mhsanaei/3x-ui/v2.svg)](https://pkg.go.dev/github.com/mhsanaei/3x-ui/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/mhsanaei/3x-ui/v2)](https://goreportcard.com/report/github.com/mhsanaei/3x-ui/v2)

**3X-UI** — продвинутая панель управления с открытым исходным кодом на основе веб-интерфейса, разработанная для управления сервером Xray-core. Предоставляет удобный интерфейс для настройки и мониторинга различных VPN и прокси-протоколов.

> [!IMPORTANT]
> Этот проект предназначен только для личного использования, пожалуйста, не используйте его в незаконных целях и в производственной среде.

Как улучшенная версия оригинального проекта X-UI, 3X-UI обеспечивает повышенную стабильность, более широкую поддержку протоколов и дополнительные функции.

## Возможности этого форка

Этот форк дополнительно включает слой управления для экосистемы TrustTunnel:

- `TrustTunnel` как отдельный управляемый протокол внутри `Inbounds`
- `MTProto` как отдельный управляемый протокол внутри `Inbounds`
- создание пользователей `TrustTunnel` прямо из панели
- экспорт `tt://` для пользователей `TrustTunnel`
- QR-экспорт для ссылок `TrustTunnel`
- автоматическую генерацию `MTProto secret`, если поле оставлено пустым

Эти доработки рассчитаны на сценарии, где используются:

- [TrustTunnel](https://github.com/TrustTunnel/TrustTunnel) как endpoint/runtime
- [trusty](https://github.com/Meddelin/trusty) и другие клиентские инструменты для GUI-подключения

На текущем этапе эти интеграции стоит воспринимать как кастомные расширения этого форка, а не как функции upstream 3X-UI.

## Быстрый старт

```bash
bash <(curl -Ls https://raw.githubusercontent.com/Valerrra/3x-ui_plus/main/install.sh)
```

После установки:

1. Откройте панель и войдите с учётными данными, которые покажет инсталлятор.
2. Установите runtime-бинарники, которые ожидает этот форк:

```bash
sudo mkdir -p /opt/trusttunnel/access /opt/trusttunnel/certs

# Установите TrustTunnel из официального проекта и убедитесь, что endpoint-бинарь доступен здесь:
sudo install -m 755 /path/to/trusttunnel_endpoint /usr/local/bin/trusttunnel_endpoint

# Установите MTProto bridge (mtg) и убедитесь, что бинарь доступен здесь:
sudo install -m 755 /path/to/mtg /usr/local/bin/mtg
```

3. Подготовьте сертификаты для TrustTunnel inbound, например так:

```bash
sudo cp /path/to/fullchain.pem /opt/trusttunnel/certs/fullchain.pem
sudo cp /path/to/privkey.pem /opt/trusttunnel/certs/privkey.pem
sudo chmod 600 /opt/trusttunnel/certs/privkey.pem
```

4. Создайте новый inbound и выберите протокол `TrustTunnel` или `MTProto`.
5. Для `TrustTunnel` используйте кнопки автозаполнения для hostname, public address и путей к сертификатам, затем сохраните inbound. Панель сама запишет:
   - `/opt/trusttunnel/vpn.toml`
   - `/opt/trusttunnel/hosts.toml`
   - `/opt/trusttunnel/credentials.toml`
6. Добавляйте пользователей `TrustTunnel` через меню inbound и сразу экспортируйте `tt://` или QR-код из панели.
7. Для `MTProto` можно оставить поле `secret` пустым, тогда панель сгенерирует корректное значение автоматически. Панель запишет `/opt/trusttunnel/access/mtproto.toml` и будет управлять `trusttunnel-mtproto.service`.

Если бинарники установлены в другие места, поправьте пути или добавьте симлинки так, чтобы форк видел:

- `/usr/local/bin/trusttunnel_endpoint`
- `/usr/local/bin/mtg`

Полную документацию смотрите в [вики проекта](https://github.com/MHSanaei/3x-ui/wiki).

## Особая благодарность

- [alireza0](https://github.com/alireza0/)

## Благодарности

- [Iran v2ray rules](https://github.com/chocolate4u/Iran-v2ray-rules) (Лицензия: **GPL-3.0**): _Улучшенные правила маршрутизации для v2ray/xray и v2ray/xray-clients со встроенными иранскими доменами и фокусом на безопасность и блокировку рекламы._
- [Russia v2ray rules](https://github.com/runetfreedom/russia-v2ray-rules-dat) (Лицензия: **GPL-3.0**): _Этот репозиторий содержит автоматически обновляемые правила маршрутизации V2Ray на основе данных о заблокированных доменах и адресах в России._

## Поддержка проекта

**Если этот проект полезен для вас, вы можете поставить ему**:star2:

<a href="https://www.buymeacoffee.com/MHSanaei" target="_blank">
<img src="./media/default-yellow.png" alt="Buy Me A Coffee" style="height: 70px !important;width: 277px !important;" >
</a>

</br>
<a href="https://nowpayments.io/donation/hsanaei" target="_blank" rel="noreferrer noopener">
   <img src="./media/donation-button-black.svg" alt="Crypto donation button by NOWPayments">
</a>

## Звезды с течением времени

[![Stargazers over time](https://starchart.cc/MHSanaei/3x-ui.svg?variant=adaptive)](https://starchart.cc/MHSanaei/3x-ui) 
