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

**3X-UI** — advanced, open-source web-based control panel designed for managing Xray-core server. It offers a user-friendly interface for configuring and monitoring various VPN and proxy protocols.

> [!IMPORTANT]
> This project is only for personal usage, please do not use it for illegal purposes, and please do not use it in a production environment.

As an enhanced fork of the original X-UI project, 3X-UI provides improved stability, broader protocol support, and additional features.

## Fork-specific additions in this repository

This fork also includes an experimental management layer for the TrustTunnel ecosystem:

- `TrustTunnel` as a custom managed protocol inside `Inbounds`
- `MTProto` as a custom managed protocol inside `Inbounds`
- `TrustTunnel` client creation directly from the panel
- `tt://` export for `TrustTunnel` users
- QR-ready export flow for `TrustTunnel` links
- automatic MTProto secret generation when the field is left empty

These additions are designed for deployments that use:

- [TrustTunnel](https://github.com/TrustTunnel/TrustTunnel) for endpoint/runtime
- [trusty](https://github.com/Meddelin/trusty) and related client-side tooling for GUI onboarding

At the moment, these fork-specific integrations should be treated as custom extensions of this repository rather than upstream 3X-UI features.

## Quick Start

```bash
bash <(curl -Ls https://raw.githubusercontent.com/Valerrra/3x-ui_plus/main/install.sh)
```

After installation:

1. Open the panel and sign in with the generated credentials.
2. Create a new inbound and choose `TrustTunnel` or `MTProto` from the protocol list.
3. For `TrustTunnel`, use the auto-fill helpers for hostname, public address, and certificate paths.
4. Add `TrustTunnel` users from the inbound menu and export `tt://` links or QR codes directly from the panel.
5. For `MTProto`, leave the secret empty to auto-generate a valid value on save.

For full documentation, please visit the [project Wiki](https://github.com/MHSanaei/3x-ui/wiki).

## A Special Thanks to

- [alireza0](https://github.com/alireza0/)

## Acknowledgment

- [Iran v2ray rules](https://github.com/chocolate4u/Iran-v2ray-rules) (License: **GPL-3.0**): _Enhanced v2ray/xray and v2ray/xray-clients routing rules with built-in Iranian domains and a focus on security and adblocking._
- [Russia v2ray rules](https://github.com/runetfreedom/russia-v2ray-rules-dat) (License: **GPL-3.0**): _This repository contains automatically updated V2Ray routing rules based on data on blocked domains and addresses in Russia._

## Support project

**If this project is helpful to you, you may wish to give it a**:star2:

<a href="https://www.buymeacoffee.com/MHSanaei" target="_blank">
<img src="./media/default-yellow.png" alt="Buy Me A Coffee" style="height: 70px !important;width: 277px !important;" >
</a>

</br>
<a href="https://nowpayments.io/donation/hsanaei" target="_blank" rel="noreferrer noopener">
   <img src="./media/donation-button-black.svg" alt="Crypto donation button by NOWPayments">
</a>

## Stargazers over Time

[![Stargazers over time](https://starchart.cc/MHSanaei/3x-ui.svg?variant=adaptive)](https://starchart.cc/MHSanaei/3x-ui)
