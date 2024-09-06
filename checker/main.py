import requests
import yaml
from bs4 import BeautifulSoup


# 从 yml 加载友链（仅限于一类友链模式）
def load_links_from_yaml(file):
    with open(file, 'r', encoding='utf-8') as file:
        data = yaml.safe_load(file)
        links = []
        for link_info in data:
            links.append(link_info['link'])
        return links


# 获取页面
def fetch_page(url):
    try:
        response = requests.get(url, timeout=10, allow_redirects=True)
        redirect_status_code = [301,302,307,308]
        if response.status_code == 200:
            return response.text
        if response.status_code in redirect_status_code:
            print(f"[info] {url} 已经被重定向到 {response.headers['location']}")
            return response.text
    except requests.RequestException as e:
        print(f"[warning] 获取 {url} 失败，原因为:\n {e}")
    return None


# 检测链接所对应的容器是否存在标题
def find_link_container(soup, target_link, link_title):
    for link in soup.find_all('a', href=True):
        if target_link in link['href']:
            container = link.find_parent()
            while container:
                if link_title in container.get_text():
                    return True
                container = container.find_parent()
    return False


# 查找友链页
def find_links_page(link):
    paths_to_check = ['/link', '/links', '/links.html', '/friend','/friends','/friends.html']
    homepage_url = link.rstrip('/')
    print(f"[info] 正在尝试访问 {homepage_url}")
    home_page_content = fetch_page(homepage_url)
    if home_page_content:
        soup = BeautifulSoup(home_page_content, 'html.parser')
        link_keywords = ['友链', '友情链接', '友情鏈接', '友人帐','朋友们', 'Links', 'Friends']

        for word in link_keywords:
            possible_link = soup.find('a', string=lambda text: text and word in text)
            if possible_link:
                target_url = possible_link.get('href')
                if target_url.startswith('javascript'):
                    continue
                    # 测试时发现有个博客写了友链，但是真正的友链在“友人帐”页面，所以加了这个判断
                elif not target_url.startswith('http'):
                    target_url = link.rstrip('/') + '/' + target_url.lstrip('/')
                    print(f"[info] 在站点首页发现'{word}'指向的友链页，链接为 {target_url}")
                return target_url
        else:
            print(f"[info] 在站点首页未能发现有效的友链页，正在尝试所有可能包含友链的页面")
            for path in paths_to_check:
                target_url = homepage_url + path
                page_content = fetch_page(target_url)
                if not page_content:
                    continue
                else:
                    print(f"[info] 在 {target_url} 发现包含友链的页面")
                    return target_url
            else:
                print(f"[warning] 在 {homepage_url} 未能发现友链页，请手动检查")
                return None


# 检查友链
def check_friend_link(link, target_link, link_title):
    target_url = find_links_page(link)
    if target_url:
        target_page_content = fetch_page(target_url)
        if target_page_content:
            target_soup = BeautifulSoup(target_page_content, 'html.parser')
            if find_link_container(target_soup, target_link, link_title):
                print(f"[info] 在 {target_url} 找到包含指定链接的的友情链接，且对应标题为 '{link_title}'")
                return
            else:
                print(f"[info] 在 {target_url} 找到包含指定链接的友情链接，但是其对应的标题似乎并不为 '{link_title}'，请手动检查")
                return
        else:
            print(f"[warning] {target_url} 似乎没有内容，需要手动介入")
            return


def main(yml_file, link, title):
    links = load_links_from_yaml(yml_file)
    for link in links:
        check_friend_link(link, link, title)


if __name__ == '__main__':
    yaml_file = './data/link.yml'  # 替换成你的 YAML 文件路径
    custom_link = 'https://lab.imgb.space'  # 替换成你要检测的链接
    keyword = '/var/log/gblab'  # 替换成你要检测的关键词
    main(yaml_file, custom_link, keyword)
