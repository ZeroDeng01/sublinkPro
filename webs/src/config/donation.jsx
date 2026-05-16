import { IconCoffee, IconGift, IconHeart, IconServer } from '@tabler/icons-react';
import affiliateData from './affiliateRecommendations.json';
import donationData from './donationLinks.json';

const donationIcons = {
  coffee: <IconCoffee size={18} />,
  gift: <IconGift size={18} />,
  heart: <IconHeart size={18} />
};

const affiliateIcon = <IconServer size={18} />;

export const affiliateRecommendationConfig = {
  title: '🌟 优质推荐',
  disclaimer:
    '此区域包含推广链接，通过这些链接购买可能会为维护者带来佣金奖励。具体价格与活动资格请以官方页面为准，我们不保证具体的延迟、网络质量或固定折扣。',
  items: affiliateData.map((item) => ({
    ...item,
    icon: affiliateIcon
  }))
};

export const donationConfig = {
  ...donationData,
  headerIcon: donationIcons[donationData.headerIcon] || donationIcons.coffee,
  links: donationData.links.map((item) => ({
    ...item,
    icon: donationIcons[item.icon] || donationIcons.coffee
  }))
};
