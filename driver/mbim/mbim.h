// SPDX-License-Identifier: MIT
/*
 * Copyright (c) 2024, Frans Klaver <frans.klaver@vislink.com>
 */

#include <libmbim-glib.h>

struct mbim_data
{
    guint32 last_channel_id;
    guint32 uim_slot;
    gboolean use_proxy;
    GMainContext *context;
    MbimDevice *device;
};

int go_mbim_apdu_connect(struct mbim_data *mbim_priv, char *device_path, char *err);
int go_mbim_apdu_disconnect(struct mbim_data *mbim_priv, char *err);
int go_mbim_apdu_transmit(struct mbim_data *mbim_priv, uint8_t **rx, uint32_t *rx_len, const uint8_t *tx, uint32_t tx_len, char *err);
int go_mbim_apdu_open_logical_channel(struct mbim_data *mbim_priv, const uint8_t *aid, uint8_t aid_len, char *err);
int go_mbim_apdu_close_logical_channel(struct mbim_data *mbim_priv, uint8_t channel, char *err);
