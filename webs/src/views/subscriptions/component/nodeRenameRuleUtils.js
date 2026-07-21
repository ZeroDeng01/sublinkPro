const QUICK_SEPARATOR_CHARS = new Set(['-', '_', '|', ' ', '[', ']', '(', ')']);

const splitSeparatorText = (text) => {
  const tokens = [];
  let buffer = '';

  for (const character of text) {
    if (QUICK_SEPARATOR_CHARS.has(character)) {
      if (buffer) {
        tokens.push(buffer);
        buffer = '';
      }
      tokens.push(character);
    } else {
      buffer += character;
    }
  }

  if (buffer) tokens.push(buffer);
  return tokens;
};

export const parseNodeRenameRule = (rule) => {
  if (!rule) return [];

  const items = [];
  let id = 0;
  const pushSeparators = (text) => {
    for (const token of splitSeparatorText(text)) {
      items.push({ id: `sep-${id++}`, type: 'separator', value: token });
    }
  };
  const variablePattern =
    /\$(Name|LinkName|LinkCountryName|LinkCountry|Flag|SpeedIcon|Speed|DelayIcon|Delay|IpType|Residential|FraudScoreIcon|FraudScore|Unlock\([^)]+\)|Unlock|Group|Source|DuplicateIndex|Index|Protocol|Tags|TagGroup\([^)]+\))/g;
  let match;
  let lastIndex = 0;

  while ((match = variablePattern.exec(rule)) !== null) {
    if (match.index > lastIndex) {
      pushSeparators(rule.substring(lastIndex, match.index));
    }
    items.push({ id: `var-${id++}`, type: 'variable', value: match[0] });
    lastIndex = match.index + match[0].length;
  }

  if (lastIndex < rule.length) {
    pushSeparators(rule.substring(lastIndex));
  }

  return items;
};

export const previewSubscriptionNodeName = (rule, previewIcons) => {
  if (!rule) return '';

  let result = rule.replace(/\$TagGroup\([^)]+\)/g, 'Fast');
  return result
    .replace(/\$Name/g, 'Hong Kong node remark')
    .replace(/\$Flag/g, '🇭🇰')
    .replace(/\$SpeedIcon/g, previewIcons.speedIcon)
    .replace(/\$DelayIcon/g, previewIcons.delayIcon)
    .replace(/\$IpType/g, 'Native IP')
    .replace(/\$Residential/g, 'Residential IP')
    .replace(/\$FraudScoreIcon/g, previewIcons.fraudScoreIcon)
    .replace(/\$FraudScore/g, '12')
    .replace(/\$Unlock\([^)]+\)/g, 'Unlock-US')
    .replace(/\$LinkName/g, 'HongKong01')
    .replace(/\$LinkCountryName/g, '香港')
    .replace(/\$LinkCountry/g, 'HK')
    .replace(/\$Speed/g, '1.50MB/s')
    .replace(/\$Delay/g, '125ms')
    .replace(/\$Group/g, 'Premium')
    .replace(/\$Source/g, 'Airport A')
    .replace(/\$DuplicateIndex/g, '1')
    .replace(/\$Index/g, '1')
    .replace(/\$Protocol/g, 'VMess')
    .replace(/\$Tags/g, 'Fast|Hong Kong');
};
