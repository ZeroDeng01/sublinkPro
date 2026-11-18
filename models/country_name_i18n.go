package models

import (
	"strings"

	"sublink/utils"
)

// countryDefaultName holds the built-in country name in both supported node
// naming languages. These values mirror the seeded country rules; when the
// stored rule name still equals one of them the name is considered untouched
// and is rendered in the requested language.
type countryDefaultName struct {
	en string
	zh string
}

var defaultCountryNames = map[string]countryDefaultName{
	"HK": {"Hong Kong", "香港"},
	"TW": {"Taiwan", "台湾"},
	"JP": {"Japan", "日本"},
	"SG": {"Singapore", "新加坡"},
	"US": {"United States", "美国"},
	"KR": {"South Korea", "韩国"},
	"MY": {"Malaysia", "马来西亚"},
	"TH": {"Thailand", "泰国"},
	"PH": {"Philippines", "菲律宾"},
	"VN": {"Vietnam", "越南"},
	"IN": {"India", "印度"},
	"ID": {"Indonesia", "印度尼西亚"},
	"BD": {"Bangladesh", "孟加拉国"},
	"PK": {"Pakistan", "巴基斯坦"},
	"MO": {"Macao", "澳门"},
	"KZ": {"Kazakhstan", "哈萨克斯坦"},
	"GB": {"United Kingdom", "英国"},
	"DE": {"Germany", "德国"},
	"FR": {"France", "法国"},
	"NL": {"Netherlands", "荷兰"},
	"RU": {"Russia", "俄罗斯"},
	"CA": {"Canada", "加拿大"},
	"BR": {"Brazil", "巴西"},
	"AR": {"Argentina", "阿根廷"},
	"MX": {"Mexico", "墨西哥"},
	"CL": {"Chile", "智利"},
	"CO": {"Colombia", "哥伦比亚"},
	"AU": {"Australia", "澳大利亚"},
	"NZ": {"New Zealand", "新西兰"},
	"TR": {"Turkey", "土耳其"},
	"AE": {"United Arab Emirates", "阿联酋"},
	"IL": {"Israel", "以色列"},
	"SA": {"Saudi Arabia", "沙特阿拉伯"},
	"QA": {"Qatar", "卡塔尔"},
	"KW": {"Kuwait", "科威特"},
	"CN": {"China", "中国"},
	"ZA": {"South Africa", "南非"},
	"EG": {"Egypt", "埃及"},
	"NG": {"Nigeria", "尼日利亚"},
	"KE": {"Kenya", "肯尼亚"},
	"IT": {"Italy", "意大利"},
	"ES": {"Spain", "西班牙"},
	"SE": {"Sweden", "瑞典"},
	"CH": {"Switzerland", "瑞士"},
	"PL": {"Poland", "波兰"},
	"PT": {"Portugal", "葡萄牙"},
	"IE": {"Ireland", "爱尔兰"},
	"NO": {"Norway", "挪威"},
	"FI": {"Finland", "芬兰"},
	"DK": {"Denmark", "丹麦"},
	"AT": {"Austria", "奥地利"},
	"BE": {"Belgium", "比利时"},
	"GR": {"Greece", "希腊"},
	"CZ": {"Czechia", "捷克"},
	"RO": {"Romania", "罗马尼亚"},
	"UA": {"Ukraine", "乌克兰"},
	"HU": {"Hungary", "匈牙利"},
	"HR": {"Croatia", "克罗地亚"},
	"RS": {"Serbia", "塞尔维亚"},
	"BG": {"Bulgaria", "保加利亚"},
	"SK": {"Slovakia", "斯洛伐克"},
	"SI": {"Slovenia", "斯洛文尼亚"},
	"LU": {"Luxembourg", "卢森堡"},
	"IS": {"Iceland", "冰岛"},
	"EE": {"Estonia", "爱沙尼亚"},
	"LV": {"Latvia", "拉脱维亚"},
	"LT": {"Lithuania", "立陶宛"},
	"BY": {"Belarus", "白俄罗斯"},
	"TJ": {"Tajikistan", "塔吉克斯坦"},
	"AQ": {"Antarctica", "南极洲"},
	"AL": {"Albania", "阿尔巴尼亚"},
	"MD": {"Moldova", "摩尔多瓦"},
	"MK": {"North Macedonia", "北马其顿"},
	"AM": {"Armenia", "亚美尼亚"},
	"UZ": {"Uzbekistan", "乌兹别克斯坦"},
	"GE": {"Georgia", "格鲁吉亚"},
	"BA": {"Bosnia and Herzegovina", "波黑"},
	"IQ": {"Iraq", "伊拉克"},
	"KH": {"Cambodia", "柬埔寨"},
	"AZ": {"Azerbaijan", "阿塞拜疆"},
	"BH": {"Bahrain", "巴林"},
	"MN": {"Mongolia", "蒙古"},
	"CY": {"Cyprus", "塞浦路斯"},
	"KG": {"Kyrgyzstan", "吉尔吉斯斯坦"},
	"BN": {"Brunei", "文莱"},
	"LI": {"Liechtenstein", "列支敦士登"},
	"LK": {"Sri Lanka", "斯里兰卡"},
	"SS": {"South Sudan", "南苏丹"},
}

// localizedDefaultCountryName resolves the built-in name for a country code,
// returning an empty string for codes without a built-in default.
func localizedDefaultCountryName(countryCode string, languageTag string) string {
	names, ok := defaultCountryNames[strings.ToUpper(strings.TrimSpace(countryCode))]
	if !ok {
		return ""
	}
	if utils.NormalizeNodeNameLanguage(languageTag) == utils.NodeNameLanguageChinese {
		return names.zh
	}
	return names.en
}

func isASCII(value string) bool {
	for _, r := range value {
		if r > 127 {
			return false
		}
	}
	return true
}

// GetLocalizedCountryName resolves the country name rendered by node naming
// variables. A stored country rule name wins unless it still equals the built-in
// default, in which case the name is rendered in the requested language. This
// keeps user-customized names intact while removing Chinese defaults from
// English output.
func GetLocalizedCountryName(countryCode string, languageTag string) string {
	stored := GetCountryNameByCode(countryCode)
	if stored == "" {
		return ""
	}

	normalizedCode := strings.ToUpper(strings.TrimSpace(countryCode))
	if names, ok := defaultCountryNames[normalizedCode]; ok {
		if stored == names.en || stored == names.zh {
			return localizedDefaultCountryName(normalizedCode, languageTag)
		}
		return stored
	}

	// Codes without a built-in default keep their stored name, except that a
	// non-ASCII name falls back to the code itself in English output.
	if utils.NormalizeNodeNameLanguage(languageTag) == utils.NodeNameLanguageEnglish && !isASCII(stored) {
		return normalizedCode
	}
	return stored
}
