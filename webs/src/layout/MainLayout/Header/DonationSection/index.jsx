import { useEffect, useRef, useState } from "react";

// material-ui
import { useTheme } from "@mui/material/styles";
import useMediaQuery from "@mui/material/useMediaQuery";
import Avatar from "@mui/material/Avatar";
import Box from "@mui/material/Box";
import ClickAwayListener from "@mui/material/ClickAwayListener";
import Divider from "@mui/material/Divider";
import Paper from "@mui/material/Paper";
import Popper from "@mui/material/Popper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import Tooltip from "@mui/material/Tooltip";
import Button from "@mui/material/Button";

// project imports
import MainCard from "ui-component/cards/MainCard";
import Transitions from "ui-component/extended/Transitions";

// assets
import { IconCoffee } from "@tabler/icons-react";

import { donationConfig } from "config/donation";

export default function DonationSection() {
  const theme = useTheme();
  const downMD = useMediaQuery(theme.breakpoints.down("md"));
  const [open, setOpen] = useState(false);
  const anchorRef = useRef(null);

  const handleToggle = () => {
    setOpen((prevOpen) => !prevOpen);
  };

  const handleClose = (event) => {
    if (anchorRef.current && anchorRef.current.contains(event.target)) {
      return;
    }
    setOpen(false);
  };

  const prevOpen = useRef(open);
  useEffect(() => {
    if (prevOpen.current === true && open === false) {
      anchorRef.current.focus();
    }
    prevOpen.current = open;
  }, [open]);

  return (
    <>
      <Box sx={{ ml: 2 }}>
        <Tooltip title="打赏支持">
          <Avatar
            variant="rounded"
            sx={{
              ...theme.typography.commonAvatar,
              ...theme.typography.mediumAvatar,
              transition: "all .2s ease-in-out",
              color: theme.palette[donationConfig.headerIconColor].dark,
              background: theme.palette[donationConfig.headerIconColor].light,
              position: "relative",
              "&:hover, &[aria-controls=\"menu-list-grow\"]": {
                color: theme.palette[donationConfig.headerIconColor].light,
                background: theme.palette[donationConfig.headerIconColor].dark
              }
            }}
            ref={anchorRef}
            aria-controls={open ? "menu-list-grow" : undefined}
            aria-haspopup="true"
            onClick={handleToggle}
          >
            <IconCoffee stroke={1.5} size="20px" />
          </Avatar>
        </Tooltip>
      </Box>
      <Popper
        placement={downMD ? "bottom" : "bottom-end"}
        open={open}
        anchorEl={anchorRef.current}
        role={undefined}
        transition
        disablePortal
        modifiers={[{ name: "offset", options: { offset: [downMD ? 5 : 0, 20] } }]}
      >
        {({ TransitionProps }) => (
          <ClickAwayListener onClickAway={handleClose}>
            <Transitions position={downMD ? "top" : "top-right"} in={open} {...TransitionProps}>
              <Paper>
                {open && (
                  <MainCard
                    border={false}
                    elevation={16}
                    content={false}
                    boxShadow
                    shadow={theme.shadows[16]}
                    sx={{ minWidth: 280 }}
                  >
                    <Stack sx={{ gap: 2, p: 2 }}>
                      <Stack direction="row" sx={{ alignItems: "center", justifyContent: "center" }}>
                        <Typography variant="subtitle1">{donationConfig.title}</Typography>
                      </Stack>
                      <Divider />
                      <Stack spacing={1.5}>
                        {donationConfig.links.map((item, index) => (
                          <Button
                            key={index}
                            variant="outlined"
                            color={item.color}
                            startIcon={item.icon}
                            href={item.url}
                            target="_blank"
                            fullWidth
                            sx={{
                              justifyContent: "flex-start",
                              px: 2,
                              py: 1.5,
                              borderRadius: 2,
                              transition: "all 0.2s",
                              "&:hover": {
                                transform: "translateY(-2px)"
                              }
                            }}
                          >
                            {item.title}
                          </Button>
                        ))}
                      </Stack>
                    </Stack>
                  </MainCard>
                )}
              </Paper>
            </Transitions>
          </ClickAwayListener>
        )}
      </Popper>
    </>
  );
}
