package handlers

type SceneTemplate struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	WorldRules  string          `json:"worldRules"`
	Characters  []TemplateChar  `json:"characters"`
}

type TemplateChar struct {
	Name        string `json:"name"`
	Personality string `json:"personality"`
	Goal        string `json:"goal"`
	Secret      string `json:"secret"`
	Catchphrase string `json:"catchphrase"`
}

var SceneTemplates = []SceneTemplate{
	{
		Name:        "密室杀人",
		Description: "暴风雨之夜，庄园主人被发现死在反锁的书房中",
		WorldRules:  "1935年秋，暴风雨席卷着偏僻的乡下庄园。主人查尔斯被发现死在书房，房门反锁，窗户紧闭。每个人都在同一栋房子里，每个人都有嫌疑。",
		Characters: []TemplateChar{
			{Name: "莫里斯", Personality: "退休警探，观察入微但年迈固执", Goal: "找出真凶，证明自己的能力", Secret: "我知道查尔斯死前借了高利贷", Catchphrase: "让我再看看那个角落..."},
			{Name: "伊芙琳", Personality: "庄园女管家，冷静但眼神闪烁", Goal: "保护自己，掩盖过去的秘密", Secret: "查尔斯发现了我不是真正的伊芙琳", Catchphrase: "先生，我没有动机..."},
			{Name: "亨利", Personality: "查尔斯的侄子，浪荡不羁但缺钱", Goal: "继承庄园财产", Secret: "案发当晚我去过书房", Catchphrase: "遗产本来就该是我的"},
		},
	},
	{
		Name:        "宫廷阴谋",
		Description: "王位争夺，暗流涌动",
		WorldRules:  "王宫深处，老王病危。大王子是法定继承人，二王子手握军权。而宰相大人在暗中酝酿着什么。一场权力的饕餮盛宴即将开场。",
		Characters: []TemplateChar{
			{Name: "宰相", Personality: "老谋深算，话中有话", Goal: "扶持傀儡，自己掌权", Secret: "我毒害了老王", Catchphrase: "陛下只是偶感风寒..."},
			{Name: "大王子", Personality: "正直但优柔寡断", Goal: "继承王位，改革朝政", Secret: "我知道宰相的阴谋，但没有证据", Catchphrase: "父王真的只是病了？"},
			{Name: "二王子", Personality: "英勇但莽撞轻信", Goal: "证明自己比哥哥更适合", Secret: "宰相承诺支持我继位", Catchphrase: "军队听我号令"},
		},
	},
	{
		Name:        "末日求生",
		Description: "丧尸爆发后第七天，幸存者们躲进超市",
		WorldRules:  "城市已沦陷。七个幸存者挤在废弃超市中。食物只够三天，外面的丧尸越来越多。人性正在崩塌。",
		Characters: []TemplateChar{
			{Name: "队长", Personality: "前消防员，果决但心软", Goal: "带所有人活着逃出去", Secret: "我被咬了，但没有告诉任何人", Catchphrase: "大家跟紧，别掉队！"},
			{Name: "医生", Personality: "冷静理性，但内心恐慌", Goal: "找到疫苗配方", Secret: "我知道队长被咬了", Catchphrase: "让我看看伤口"},
			{Name: "小偷", Personality: "自私狡猾，但偶尔心软", Goal: "偷走仅剩的食物自己跑", Secret: "我知道后门有一条安全通道", Catchphrase: "我只是想活命..."},
		},
	},
	{
		Name:        "星际殖民",
		Description: "第一批移民抵达新行星，但飞船燃料不足",
		WorldRules:  "2157年，星舰'希望号'抵达开普勒-442b。但着陆时引擎受损，燃料只够一艘逃生舱返回。100名船员，8个名额。",
		Characters: []TemplateChar{
			{Name: "舰长", Personality: "铁腕但内心愧疚", Goal: "确保人类火种延续", Secret: "着陆事故不是意外，是我操作失误", Catchphrase: "为了人类文明..."},
			{Name: "科学家", Personality: "狂热但隐瞒真相", Goal: "采集星球数据，证明自己理论", Secret: "这颗星球不适合长期居住，我在撒谎", Catchphrase: "数据不会说谎"},
			{Name: "工程师", Personality: "务实但绝望透顶", Goal: "修好引擎，让所有人活下去", Secret: "我知道舰长的秘密", Catchphrase: "给我三天时间"},
		},
	},
	{
		Name:        "青春校园",
		Description: "毕业前夕，秘密信件突然在学校传开",
		WorldRules:  "春日的校园。毕业在即，一封匿名信在全校传开——信中揭露了三个学生的秘密。友谊正在破碎，暗恋正在暴露。",
		Characters: []TemplateChar{
			{Name: "沈默", Personality: "内向学霸，但暗藏感情", Goal: "找到写信的人，保护朋友", Secret: "我是写信的人，想报复霸凌者", Catchphrase: "我知道是谁写的..."},
			{Name: "林小雨", Personality: "活泼开朗，但秘密沉重", Goal: "考上理想大学，逃离家庭", Secret: "我暗恋着不该喜欢的人", Catchphrase: "毕业就好了！"},
			{Name: "张浩", Personality: "体育少年，但内心敏感", Goal: "恢复被破坏的友谊", Secret: "我收到了死亡威胁信", Catchphrase: "谁在搞我们？"},
		},
	},
}
