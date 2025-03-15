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

MbimDevice *
mbim_device_new_from_path(
    GFile *file,
    GMainContext *context,
    GError **error);

gboolean
mbim_device_open_sync(
    MbimDevice *device,
    MbimDeviceOpenFlags open_flags,
    GMainContext *context,
    GError **error);

gboolean
mbim_device_close_sync(
    MbimDevice *device,
    GMainContext *context,
    GError **error);

MbimMessage *
mbim_device_command_sync(
    MbimDevice *device,
    GMainContext *context,
    MbimMessage *request,
    GError **error);

int go_mbim_apdu_connect(struct mbim_data *mbim_priv, char *device_path);
void go_mbim_apdu_disconnect(struct mbim_data *mbim_priv);
int go_mbim_apdu_transmit(struct mbim_data *mbim_priv, uint8_t **rx, uint32_t *rx_len, const uint8_t *tx, uint32_t tx_len);
int go_mbim_apdu_open_logical_channel(struct mbim_data *mbim_priv, const uint8_t *aid, uint8_t aid_len);
int go_mbim_apdu_close_logical_channel(struct mbim_data *mbim_priv, uint8_t channel);
