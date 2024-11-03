
- 初始化用户本地环境
	- 帮助用户生成一个新的系统用户
	- 用户创建站点
		- 帮助用户生成Author
			- 一个是主题Author: mdfriday.com。 生成前查询是否已经创建。
			- 一个是用户自己：随机first name, last name。创建后记录在配置信息中。
		- 注册用户选择的主题
			- 查询主题是否已经存在
		- 创建站点

### 创建用户

curl -X POST http://127.0.0.1:1314/api/user \
-H "Content-Type: application/x-www-form-urlencoded" \
-d "email=abc@qq.com&password=123456"

curl -X POST http://127.0.0.1:1314/api/login \
-H "Content-Type: application/x-www-form-urlencoded" \
-d "email=me@sunwei.xyz&password=123456"

### 创建站点

curl -X POST "http://127.0.0.1:1314/api/content?type=Site" \
-H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOm51bGwsImV4cCI6IjIwMjQtMTItMDNUMTI6MjM6NTMuODY2MTY3KzA4OjAwIiwiaWF0IjpudWxsLCJpc3MiOm51bGwsImp0aSI6bnVsbCwibmJmIjpudWxsLCJzdWIiOm51bGwsInVzZXIiOiJtZUBzdW53ZWkueHl6In0.2xU4XR12r_2QxQ1z3dKy7WZr2qyZwHLYi_jN8OnjyXg" \
-F "type=Site" \
-F "title=Demo" \
-F "description=This is my first demo site created by hugoverse" \
-F "base_url=/" \
-F "theme=github.com/mdfriday/theme-manual-of-me" \
-F "owner=me@sunwei.xyz" \
-F "Params=Author = '老袁讲敏捷'
CoverImage = 'cover.jpeg'" \
-F "working_dir=/.local/share/temp"


#### 站点语言，可选，默认EN


#### 创建Post

curl -X POST "http://127.0.0.1:1314/api/content?type=Post" \
-H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOm51bGwsImV4cCI6IjIwMjQtMTItMDNUMTI6MjM6NTMuODY2MTY3KzA4OjAwIiwiaWF0IjpudWxsLCJpc3MiOm51bGwsImp0aSI6bnVsbCwibmJmIjpudWxsLCJzdWIiOm51bGwsInVzZXIiOiJtZUBzdW53ZWkueHl6In0.2xU4XR12r_2QxQ1z3dKy7WZr2qyZwHLYi_jN8OnjyXg" \
-F "type=Post" \
-F "title=关于我" \
-F "author=laoyuan" \
-F "params=weight: 1" \
-F "assets.0=@/Users/sunwei/Downloads/good.jpeg" \
-F "content=- **个人长期陪跑教练**
- 企业级敏捷教练
- 研发团队效能顾问
- unFIX中文社区发起人
- 中国最大的敏捷主题个人自媒体（bilibili \"老袁讲敏捷\"）
- \"老袁讲敏捷\" 公众号和视频号
- 长篇小说作家（湖北省作协会员）

\![good](good.jpeg)

---

"


创建SitePost

curl -X POST "http://127.0.0.1:1314/api/content?type=SitePost" \
-H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOm51bGwsImV4cCI6IjIwMjQtMTItMDNUMTI6MjM6NTMuODY2MTY3KzA4OjAwIiwiaWF0IjpudWxsLCJpc3MiOm51bGwsImp0aSI6bnVsbCwibmJmIjpudWxsLCJzdWIiOm51bGwsInVzZXIiOiJtZUBzdW53ZWkueHl6In0.2xU4XR12r_2QxQ1z3dKy7WZr2qyZwHLYi_jN8OnjyXg" \
-F "site=/api/content?type=Site&id=13" \
-F "post=/api/content?type=Post&id=17" \
-F "path=/content/01.service.md"

#### Preview

curl -X POST "http://127.0.0.1:1314/api/preview?type=Site&id=13" \
-H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOm51bGwsImV4cCI6IjIwMjQtMTItMDNUMTI6MjM6NTMuODY2MTY3KzA4OjAwIiwiaWF0IjpudWxsLCJpc3MiOm51bGwsImp0aSI6bnVsbCwibmJmIjpudWxsLCJzdWIiOm51bGwsInVzZXIiOiJtZUBzdW53ZWkueHl6In0.2xU4XR12r_2QxQ1z3dKy7WZr2qyZwHLYi_jN8OnjyXg"


#### Deployment

curl -X POST "http://127.0.0.1:1314/api/deploy?type=Site&id=13" \
-H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOm51bGwsImV4cCI6IjIwMjQtMTItMDNUMDk6NTA6NDEuOTEwMDE3KzA4OjAwIiwiaWF0IjpudWxsLCJpc3MiOm51bGwsImp0aSI6bnVsbCwibmJmIjpudWxsLCJzdWIiOm51bGwsInVzZXIiOiJtZUBzdW53ZWkueHl6In0.UISiT9zdJS1KQDT_K6o81jPybBdxz51952JGTZmYkhs"



#### Search

curl -X GET "http://127.0.0.1:1314/api/search?type=SitePost&q=site:%2Fapi%2Fcontent%3Ftype%3DSite%26id%3D9" \
-H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOm51bGwsImV4cCI6IjIwMjQtMTItMDNUMDk6NTA6NDEuOTEwMDE3KzA4OjAwIiwiaWF0IjpudWxsLCJpc3MiOm51bGwsImp0aSI6bnVsbCwibmJmIjpudWxsLCJzdWIiOm51bGwsInVzZXIiOiJtZUBzdW53ZWkueHl6In0.UISiT9zdJS1KQDT_K6o81jPybBdxz51952JGTZmYkhs"
