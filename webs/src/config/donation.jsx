import { IconCoffee, IconGift, IconHeart, IconServer, IconCloud } from '@tabler/icons-react';
import affiliateData from './affiliateRecommendations.json';

export const affiliateRecommendationConfig = {
  title: '🌟 优质推荐',
  disclaimer:
    '此区域包含推广链接，通过这些链接购买可能会为维护者带来佣金奖励。具体价格与活动资格请以官方页面为准，我们不保证具体的延迟、网络质量或固定折扣。',
  items: affiliateData.map((item) => ({
    ...item,
    icon: item.id === 'bandwagonhost' ? <IconServer size={18} /> : <IconCloud size={18} />
  }))
};

export const donationConfig = {
  headerIconColor: 'primary',
  title: '💖 感谢支持',
  links: [
    {
      title: '低调佬友打赏',
      url: 'https://credit.linux.do/paying/online?token=d03d70e9fde196dc2653a27da7a82153108ff4ae42562059714065471d7bdaea',
      icon: <IconCoffee size={18} />,
      color: 'primary'
    },
    {
      title: '豪气佬友打赏',
      url: 'https://credit.linux.do/paying/online?token=b56b0e07002b9242bcedde7947820e36970e29156d3250b0aa8c0905dd4fcf9a',
      icon: <IconGift size={18} />,
      color: 'success'
    },
    {
      title: '豪气佬友专项扶贫',
      url: 'https://credit.linux.do/paying/online?token=22a34921d096fb1c0eb837d4467ac58fc24a6c78ab6200528374a2058fc8ccf9',
      icon: <IconHeart size={18} />,
      color: 'error'
    }
  ]
};
