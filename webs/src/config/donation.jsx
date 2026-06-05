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
  titleKey: 'affiliate.title',
  disclaimerKey: 'affiliate.disclaimer',
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
