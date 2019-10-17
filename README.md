# cloud-plan-migrate

## Overview

石狩第1ゾーンでの旧プランのサーバ/ディスクを新プランに移行します。  

### 処理の流れ

- サーバが起動している場合はシャットダウンする
- ディスクの切断
- 各ディスクをクローンし新プランのディスクを作成
- ディスクの接続
- サーバのプラン変更
- サーバ起動(デフォルト:有効、オプションで無効化可能)
- 旧ディスク削除(デフォルト:無効、オプション指定時のみ有効)

## Install

リリースページから最新の実行ファイルをダウンロードして展開してください。

[https://github.com/sacloud/cloud-plan-migrate/releases/latest](https://github.com/sacloud/cloud-plan-migrate/releases/latest)

## Usage

実行するにはAPIキー(トークン/シークレット)と対象サーバのID or 名称が必要です。  
APIキーのアクセスレベルは[`作成・削除`](https://manual.sakura.ad.jp/cloud/controlpanel/access-level.html)が必要です。

APIキーと対象サーバのID or 名称が準備できたら以下のように実行します。  

```bash
$ cloud-plan-migrate <ID or 名称>

Your API AccessToken is not set
	Enter your token: <トークンを入力>
Your API AccessTokenSecret is not set
	Eneter your secret: <シークレットを入力>
```

処理対象はサーバのIDまたは名称を指定します。(スペース区切りで複数指定可)  
または、`--selector`オプションで対象リソースをタグで指定することも可能です。

実行するとカレントディレクトリ配下に`migrate-[yyyyMMdd-HHmmss].log`という名称のログファイルが出力されます。  
(`[yyyyMMdd-HHmmss]`部分は現在日時となります)

もし処理対象に以下のサーバが含まれている場合はエラーとなります。

- ディスクが接続されていない場合
- 既に新プランに移行している場合

APIキーは環境変数、またはコマンドラインオプションでも指定可能です。  
詳細は [Options](#Options)を参照してください。  

### Options

オプションは以下のように引数の前に指定します。

```bash
# --cleanup-diskオプションを指定する例
$ cloud-plan-migrate --cleanup-disk <ID or 名称>

# --selectorオプションを指定する例
$ cloud-plan-migrate --selector <処理したいサーバのタグ>
```

以下のオプションが指定可能です。

#### APIキー関連

- `--token`: [必須] APIキー(トークン)、環境変数 `SAKURACLOUD_ACCESS_TOKEN`でも指定可能です。
- `--secret`: [必須] APIキー(シークレット)、環境変数 `SAKURACLOUD_ACCESS_TOKEN_SECRET`でも指定可能です。

これらを省略した場合、実行時に入力を促すダイアログが表示されます。  

#### マイグレーションの動作関連

- `--selector`: 対象サーバをタグで指定する
- `--disable-reboot`: プラン変更後にサーバの起動を行わない
- `--cleanup-disk`: プラン変更後に旧ディスクを削除する

### その他

- `--assumeyes/-y`: 実行前の確認を省略する

## Dockerで実行する場合

cloud-plan-migrateはDockerイメージも提供しています。

[https://hub.docker.com/r/sacloud/cloud-plan-migrate/](https://hub.docker.com/r/sacloud/cloud-plan-migrate/)

以下のように利用します。

```bash
# カレントディレクトリにログ出力されるため-vオプションを指定
$ docker run -it --rm -v $PWD:/work sacloud/cloud-plan-migrate <ID or 名称>
```

## 注意/制限事項

- 旧プランのディスクにおいて`標準プラン:20GB`プランを利用していた場合、新プランに対応するプランが存在しないため`SSDプラン:20GB`プランへと変更されます。
- 対象サーバはACPIに対応している必要があります。サーバがACPI非対応の場合、当ツールからの電源オフ操作が行えずタイムアウトエラーとなり処理が中断されます。この場合、ツール実行前に手動で電源オフ操作を行ってください。  
- 同時に処理できる上限はサーバ10台分です。10台以上を処理する場合、10台を超える部分については前の処理が終わり次第逐次処理されます。  
- ディスク作成時の[ストレージ分散](https://manual.sakura.ad.jp/cloud/storage/disk.html#id3)の指定には非対応です。  

## License

 `cloud-plan-migrate` Copyright (C) 2019 Kazumichi Yamamoto.

  This project is published under [Apache 2.0 License](LICENSE.txt).
  
## Author

  * [Kazumichi Yamamoto](https://github.com/yamamoto-febc)
