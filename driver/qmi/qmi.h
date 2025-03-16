// SPDX-License-Identifier: MIT
/*
 * Copyright (c) 2024, Robert Marko <robert.marko@sartura.hr>
 */

#include <stdint.h>

#include <libqmi-glib.h>

struct qmi_data
{
    guint32 last_channel_id;
    guint32 uim_slot;
    gboolean use_proxy;
    GMainContext *context;
    QmiClientUim *uim_client;
};

int go_qmi_apdu_connect(struct qmi_data *qmi_priv, char *device_path, char *err);
int go_qmi_apdu_disconnect(struct qmi_data *qmi_priv, char *err);
int go_qmi_apdu_transmit(struct qmi_data *qmi_priv, uint8_t **rx, uint32_t *rx_len, const uint8_t *tx, uint32_t tx_len, char *err);
int go_qmi_apdu_open_logical_channel(struct qmi_data *qmi_priv, const uint8_t *aid, uint8_t aid_len, char *err);
int go_qmi_apdu_close_logical_channel(struct qmi_data *qmi_priv, uint8_t channel, char *err);
