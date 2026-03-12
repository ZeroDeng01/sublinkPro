import { IconCoffee, IconGift, IconHeart } from "@tabler/icons-react";

export const donationConfig = {
  headerIconColor: "primary",
  title: "💖 感谢支持",
  links: [
    {
      title: "低调佬友打赏",
      url: "https://credit.linux.do/paying/online?token=d03d70e9fde196dc2653a27da7a82153108ff4ae42562059714065471d7bdaea",
      icon: <IconCoffee size={18} />,
      color: "primary"
    },
    {
      title: "豪气佬友打赏",
      url: "https://credit.linux.do/paying/online?token=b56b0e07002b9242bcedde7947820e36970e29156d3250b0aa8c0905dd4fcf9a",
      icon: <IconGift size={18} />,
      color: "success"
    },
    {
      title: "豪气佬友专项扶贫",
      url: "https://credit.linux.do/paying/online?token=22a34921d096fb1c0eb837d4467ac58fc24a6c78ab6200528374a2058fc8ccf9",
      icon: <IconHeart size={18} />,
      color: "error"
    }
  ]
};
