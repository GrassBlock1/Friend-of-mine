# 草方块的朋友们
哦，你也想成为咱的朋友吗？

![GitHub Repo stars](https://img.shields.io/github/stars/Grassblock1/Friend-of-mine?style=for-the-badge)
![GitHub forks](https://img.shields.io/github/forks/GrassBlock1/friend-of-mine?color=green&style=for-the-badge)

## Warning
此处提交的是"Mare_Infinitus"(lab.imgb.space)的友链信息。

**本库正式转为提交更改YAML的方式进行收集友链，请不要再对旧有的JSON文件做任何修改！！**
## Thinking
在这个独立博客式微的时代，友情链接的意义是什么？
按照 [Wikipedia](https://zh.wikipedia.org/wiki/%E5%8F%8B%E6%83%85%E9%93%BE%E6%8E%A5) 上的定义，这种「两个或者以上的网站互相放上对方的链接，达到向用户推荐以互相共享用户群的一种 SEO 方式」有着不少优点，包括但不限于提高网站权重、知名度、转化率。
和 Medium、简书之类的写作平台不同，每一个独立博客都是一个信息孤岛。我们没有类似「你可能喜欢其他人写的这些文章」的推荐机制，所以除了搜索引擎、社交网络引荐之外，我们应该还需要通过某种方法将这些信息孤岛连接起来：交换友情链接就是一种很棒的架桥方式。

当然，我在意的不只是所谓的推广，更多的是想认识更多的人，看看更大的世界，学到更多知识。


/* 以上内容部分来自 [这儿](https://printempw.github.io/friends/) */

## Applying
### 如果你目前没有一个博客/个人主页
**仅限比较眼熟的人开这类请求，而且最好自证一下身份
1. 你可以建立一个。如果你需要帮助，欢迎开[issue](https://github.com/GrassBlock1/Friend-of-mine/issues)。
2. 暂时没有建立的想法？你可以提交常用的社交平台地址，并在PR中说明。
3. 如果实在没有，你可以只 Fork 这个repo。

### 如果你有一个博客/个人主页
要求如下：
1. 首先，友链友链，先友后链嘛。所以最好是草方块比较眼熟的人呢。
2. 关于域名
    - 可以接受免费域名，但是限制在[这个列表](https://github.com/GrassBlock1/Friend-of-mine/blob/master/url-whitelist.txt)中的**部分域名**
    - 不接受包含特殊反面含义的域名...（像什么 fxxkxxx.cn 什么的）
3. 站点启用了HTTPS， 2022年都快过去了，这条不过分吧
4. 尽可能保证SLA最大化，如果确实由于意外情况需要关站等等请开issue说明。
5. 不存在浏览器限制，比如仅限某一浏览器浏览，但是**存有“推荐使用现代浏览器浏览”的提示的网站不计入此类。**
6. 内容上尽量原创/标明来源搬运，有一定**实质内容**，不接受采集站、存在严重洗稿的网站。
7. 网站需要人类可读，~~没有css也行，但是不能在该有css的地方让css缺席~~（（
8. 网站需要宽容ADblocker等插件，不接受使用Anti-Adblocker等服务使得访客无法正常存取网站的网站（当然有诸如“为了更好支持站点，请关闭广告屏蔽插件”的提示，但不影响正常访问的可以接受）

无论如何，你都需要掌握基本的Git&Github使用方法，以及一点YAML知识 ~~（其实没有的话也行，用VS Code什么的自动补全）~~ 。

### How to
#### 如果你有一个博客/个人主页
1. 添加本站信息

>    名字：Mare_Infinitus
>
>    站长：草方块
>
>    链接：https://lab.imgb.space
>
>    介绍：77569号奇点观察员的观察日志
>
>    Logo：https://lab.imgb.space/img/favicon@square.png
>
>    横幅：还没做.jpg

*介绍可以自定义的说*

*如果你的友链页面没有放图片的地方、就不用管 Logo 了，没关系的说~*

2. 准备自己站点的信息
    - 站点名称不要超过25个半角字符，否则会在展示时被截断（*即使hover后，能显示的字符只有20个全角字符*）
    - 站点介绍不要超过13个半角字符，否则会在展示时被截断
    - 站点Logo的要求：
        - 中心对称图形，如正方形、圆形、菱形等
        - 长度及宽度不超过1440px
        - 建议使用长期稳定的托管服务托管logo，如果实在没有可以将logo放在 img 文件夹下
        - 使用现代浏览器（如 Chrome、Firefox、Opera等等）可以正常查看的文件格式，如 `png`、`jpg`、`webp`、`avif`、`svg`、`ico` 等等
    - 原则上站点信息应当适合在任何网站上展示给任何年龄段的任何人
3. 在Github上Fork此仓库。
4. 参照 [Butterfly 文档](https://butterfly.js.org/posts/dc584b87/#%E5%8F%8B%E6%83%85%E9%8F%88%E6%8E%A5%E6%B7%BB%E5%8A%A0) 和[YAML 入门](https://www.runoob.com/w3cnote/yaml-intro.html)修改 `data/link.yml` ~~或者直接参照已有格式添加就可以~~

    格式大致如下（仅供参考）：
    ```yaml
        - name: "站点名称",
          link: "https://example.com/",
          avatar: "https://example.com/img.png",
          descr: "站点介绍"  
    ```
5. 完成后新建一个Pull Request即可。
当你的Pull Request被通过后，即可显示在[友链页面](https://lab.imgb.space/friends)。

### 如果你目前没有一个博客/个人主页
参考[如果你目前没有一个博客/个人主页](#如果你目前没有一个博客个人主页)

如果博客建立完成则按照上面的步骤进行。

如果是社交平台请尽量提及一下咱在对应平台的账号（？
