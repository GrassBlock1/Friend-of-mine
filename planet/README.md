# 星球

“星球”是一种将各个网站的 RSS 进行聚合以便阅读的方式。

## 站点

https://planet.lab.gb0.dev

### 用到的项目

- [kgaughan/mercury: A Planet-style feed aggregator](https://github.com/kgaughan/mercury)
- GitHub Pages

## 对于访问者

点击标题即可跳转到源站阅读文章，网站默认展示前 30 篇文章。

如果要订阅所有站点的更新，可以通过以下 RSS 地址订阅：

```
https://planet.lab.gb0.dev/atom.xml
```

或者使用 OPML：

```
https://planet.lab.gb0.dev/opml.xml
```

## 对于站长

提交友链的同时修改 `mercury.toml`，按格式添加即可。

如果你的文章在 PR 被合并之后并没有出现在列表中，可能有以下原因：

- 上次更新文章的时间过于久远，导致无法被收入最近文章当中
- RSS 的获取被 Cloudflare / 其它提供商的防止自动程序的机制拦截了，请为 RSS URL 添加一条对于 UA 包含 `planet-mercury` 放行的规则。
- RSS 格式过于不规范，导致程序解析失败了

如果你使用 Cloudflare，并且 RSS 获取因此失败了，请在域名下面的 Security > Security Rule 中添加新的规则：

```
(http.request.method eq "GET" and http.request.uri.path eq "/rss.xml" and http.user_agent contains "planet-mercury")
```

并在操作当中选择 Skip，并尝试勾选：
```
All managed rules
All Super Bot Fight Mode Rules
```

如果还不行，请尝试关闭 `Bot Fight Mode` 之后重试。

## 对于开发

请先行按照[文档](https://kgaughan.github.io/mercury/)安装 `mercury` ，要注意的是，如果你使用新版本的 Golang ，安装命令应当为：

```shell
go install github.com/kgaughan/mercury/cmd/mercury@latest
```

然后直接在这个目录运行 mercury 即可生成对应的网页到 `./output` 目录，可以使用任何一个 http server 作为服务器来预览这个目录。

主题在 `./themes/default` 目录，按照 [Themes - Planet Mercury](https://kgaughan.github.io/mercury/themes/) 和 GoLang  的 HTML 模板语法进行修改即可。

要注意每次修改都需要重新构建一次站点，建议通过脚本来自动化以便开发。

