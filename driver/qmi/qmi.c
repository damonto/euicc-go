// SPDX-License-Identifier: MIT
/*
 * Copyright (c) 2024, Robert Marko <robert.marko@sartura.hr>
 */

#include <stdlib.h>
#include <stdio.h>

#include "qmi.h"

static void
async_result_ready(GObject *source_object,
                   GAsyncResult *res,
                   gpointer user_data)
{
    GAsyncResult **result_out = user_data;

    g_assert(*result_out == NULL);
    *result_out = g_object_ref(res);
}

QmiDevice *
qmi_device_new_from_path(GFile *file,
                         GMainContext *context,
                         GError **error)
{
    g_autoptr(GMainContextPusher) pusher = NULL;
    g_autoptr(GAsyncResult) result = NULL;
    g_autofree gchar *id = NULL;

    pusher = g_main_context_pusher_new(context);

    id = g_file_get_path(file);
    if (id)
        qmi_device_new(file,
                       NULL,
                       async_result_ready,
                       &result);

    while (result == NULL)
        g_main_context_iteration(context, TRUE);

    return qmi_device_new_finish(result, error);
}

gboolean
qmi_device_open_sync(QmiDevice *device,
                     QmiDeviceOpenFlags open_flags,
                     GMainContext *context,
                     GError **error)
{
    g_autoptr(GMainContextPusher) pusher = NULL;
    g_autoptr(GAsyncResult) result = NULL;

    pusher = g_main_context_pusher_new(context);

    qmi_device_open(device,
                    open_flags,
                    15,
                    NULL,
                    async_result_ready,
                    &result);

    while (result == NULL)
        g_main_context_iteration(context, TRUE);

    return qmi_device_open_finish(device, result, error);
}

QmiClient *
qmi_device_allocate_client_sync(QmiDevice *device,
                                GMainContext *context,
                                GError **error)
{
    g_autoptr(GMainContextPusher) pusher = NULL;
    g_autoptr(GAsyncResult) result = NULL;

    pusher = g_main_context_pusher_new(context);

    qmi_device_allocate_client(device,
                               QMI_SERVICE_UIM,
                               QMI_CID_NONE,
                               10,
                               NULL,
                               async_result_ready,
                               &result);

    while (result == NULL)
        g_main_context_iteration(context, TRUE);

    return qmi_device_allocate_client_finish(device, result, error);
}

gboolean
qmi_device_release_client_sync(QmiDevice *device,
                               QmiClient *client,
                               GMainContext *context,
                               GError **error)
{
    g_autoptr(GMainContextPusher) pusher = NULL;
    g_autoptr(GAsyncResult) result = NULL;

    pusher = g_main_context_pusher_new(context);

    qmi_device_release_client(device,
                              client,
                              QMI_DEVICE_RELEASE_CLIENT_FLAGS_RELEASE_CID,
                              10,
                              NULL,
                              async_result_ready,
                              &result);

    while (result == NULL)
        g_main_context_iteration(context, TRUE);

    return qmi_device_release_client_finish(device, result, error);
}

QmiMessageUimOpenLogicalChannelOutput *
qmi_client_uim_open_logical_channel_sync(
    QmiClientUim *client,
    QmiMessageUimOpenLogicalChannelInput *input,
    GMainContext *context,
    GError **error)
{
    g_autoptr(GMainContextPusher) pusher = NULL;
    g_autoptr(GAsyncResult) result = NULL;

    pusher = g_main_context_pusher_new(context);

    qmi_client_uim_open_logical_channel(client,
                                        input,
                                        10,
                                        NULL,
                                        async_result_ready,
                                        &result);

    while (result == NULL)
        g_main_context_iteration(context, TRUE);

    return qmi_client_uim_open_logical_channel_finish(client, result, error);
}

QmiMessageUimLogicalChannelOutput *
qmi_client_uim_logical_channel_sync(
    QmiClientUim *client,
    QmiMessageUimLogicalChannelInput *input,
    GMainContext *context,
    GError **error)
{
    g_autoptr(GMainContextPusher) pusher = NULL;
    g_autoptr(GAsyncResult) result = NULL;

    pusher = g_main_context_pusher_new(context);

    qmi_client_uim_logical_channel(client,
                                   input,
                                   10,
                                   NULL,
                                   async_result_ready,
                                   &result);

    while (result == NULL)
        g_main_context_iteration(context, TRUE);

    return qmi_client_uim_logical_channel_finish(client, result, error);
}

QmiMessageUimSendApduOutput *
qmi_client_uim_send_apdu_sync(
    QmiClientUim *client,
    QmiMessageUimSendApduInput *input,
    GMainContext *context,
    GError **error)
{
    g_autoptr(GMainContextPusher) pusher = NULL;
    g_autoptr(GAsyncResult) result = NULL;

    pusher = g_main_context_pusher_new(context);

    qmi_client_uim_send_apdu(client,
                             input,
                             10,
                             NULL,
                             async_result_ready,
                             &result);

    while (result == NULL)
        g_main_context_iteration(context, TRUE);

    return qmi_client_uim_send_apdu_finish(client, result, error);
}

int go_qmi_apdu_connect(struct qmi_data *qmi_priv, char *device_path, char *err)
{
    g_autoptr(GError) error = NULL;
    QmiDevice *device = NULL;
    QmiClient *client = NULL;
    GFile *file;

    file = g_file_new_for_path(device_path);

    qmi_priv->context = g_main_context_new();
    device = qmi_device_new_from_path(file, qmi_priv->context, &error);
    if (!device)
    {
        strncpy(err, error->message, strlen(error->message));
        return -1;
    }

    QmiDeviceOpenFlags open_flags = QMI_DEVICE_OPEN_FLAGS_NONE;
    if (qmi_priv->use_proxy)
        open_flags |= QMI_DEVICE_OPEN_FLAGS_PROXY;

    qmi_device_open_sync(device, open_flags, qmi_priv->context, &error);
    if (error)
    {
        strncpy(err, error->message, strlen(error->message));
        return -1;
    }

    client = qmi_device_allocate_client_sync(device, qmi_priv->context, &error);
    if (!client)
    {
        strncpy(err, error->message, strlen(error->message));
        return -1;
    }

    qmi_priv->uim_client = QMI_CLIENT_UIM(client);

    return 0;
}

int go_qmi_apdu_disconnect(struct qmi_data *qmi_priv, char *err)
{
    int ret = 0;
    g_autoptr(GError) error = NULL;
    QmiClient *client = QMI_CLIENT(qmi_priv->uim_client);
    QmiDevice *device = QMI_DEVICE(qmi_client_get_device(client));

    qmi_device_release_client_sync(device, client, qmi_priv->context, &error);
    qmi_priv->uim_client = NULL;

    if (error)
    {
        ret = -1;
        strncpy(err, error->message, strlen(error->message));
    }

    g_main_context_unref(qmi_priv->context);
    qmi_priv->context = NULL;

    if (qmi_priv->last_channel_id > 0)
    {
        go_qmi_apdu_close_logical_channel(qmi_priv, qmi_priv->last_channel_id, err);
        qmi_priv->last_channel_id = 0;
    }

    qmi_priv->last_channel_id = 0;
    qmi_priv->uim_slot = 0;

    return ret;
}

int go_qmi_apdu_transmit(struct qmi_data *qmi_priv, uint8_t **rx, uint32_t *rx_len, const uint8_t *tx, uint32_t tx_len, char *err)
{
    g_autoptr(GError) error = NULL;
    g_autoptr(GArray) apdu_data = NULL;

    /* Convert tx into request GArray */
    apdu_data = g_array_new(FALSE, FALSE, sizeof(guint8));
    for (uint32_t i = 0; i < tx_len; i++)
        g_array_append_val(apdu_data, tx[i]);

    QmiMessageUimSendApduInput *input;
    input = qmi_message_uim_send_apdu_input_new();
    qmi_message_uim_send_apdu_input_set_slot(input, qmi_priv->uim_slot, NULL);
    qmi_message_uim_send_apdu_input_set_channel_id(input, qmi_priv->last_channel_id, NULL);
    qmi_message_uim_send_apdu_input_set_apdu(input, apdu_data, NULL);

    QmiMessageUimSendApduOutput *output;
    output = qmi_client_uim_send_apdu_sync(qmi_priv->uim_client, input, qmi_priv->context, &error);

    qmi_message_uim_send_apdu_input_unref(input);

    if (!qmi_message_uim_send_apdu_output_get_result(output, &error))
    {
        strncpy(err, error->message, strlen(error->message));
        return -1;
    }

    GArray *apdu_res = NULL;
    if (!qmi_message_uim_send_apdu_output_get_apdu_response(output, &apdu_res, &error))
    {
        strncpy(err, error->message, strlen(error->message));
        return -1;
    }

    /* Convert response GArray into rx */
    *rx_len = apdu_res->len;
    *rx = malloc(*rx_len);
    if (!*rx)
        return -1;
    for (guint i = 0; i < apdu_res->len; i++)
        (*rx)[i] = apdu_res->data[i];

    qmi_message_uim_send_apdu_output_unref(output);

    return 0;
}

int go_qmi_apdu_open_logical_channel(struct qmi_data *qmi_priv, const uint8_t *aid, uint8_t aid_len, char *err)
{
    g_autoptr(GError) error = NULL;
    guint8 channel_id;

    GArray *aid_data = g_array_new(FALSE, FALSE, sizeof(guint8));
    for (int i = 0; i < aid_len; i++)
        g_array_append_val(aid_data, aid[i]);

    QmiMessageUimOpenLogicalChannelInput *input;
    input = qmi_message_uim_open_logical_channel_input_new();
    qmi_message_uim_open_logical_channel_input_set_slot(input, qmi_priv->uim_slot, NULL);
    qmi_message_uim_open_logical_channel_input_set_aid(input, aid_data, NULL);

    QmiMessageUimOpenLogicalChannelOutput *output;
    output = qmi_client_uim_open_logical_channel_sync(qmi_priv->uim_client, input, qmi_priv->context, &error);

    qmi_message_uim_open_logical_channel_input_unref(input);
    g_array_unref(aid_data);

    if (!output)
    {
        strncpy(err, error->message, strlen(error->message));
        return -1;
    }

    if (!qmi_message_uim_open_logical_channel_output_get_result(output, &error))
    {
        strncpy(err, error->message, strlen(error->message));
        return -1;
    }

    if (!qmi_message_uim_open_logical_channel_output_get_channel_id(output, &channel_id, &error))
    {
        strncpy(err, error->message, strlen(error->message));
        return -1;
    }
    qmi_priv->last_channel_id = channel_id;

    qmi_message_uim_open_logical_channel_output_unref(output);

    return channel_id;
}

int go_qmi_apdu_close_logical_channel(struct qmi_data *qmi_priv, uint8_t channel, char *err)
{
    g_autoptr(GError) error = NULL;

    QmiMessageUimLogicalChannelInput *input;
    input = qmi_message_uim_logical_channel_input_new();
    qmi_message_uim_logical_channel_input_set_slot(input, qmi_priv->uim_slot, NULL);
    qmi_message_uim_logical_channel_input_set_channel_id(input, channel, NULL);

    QmiMessageUimLogicalChannelOutput *output;
    output = qmi_client_uim_logical_channel_sync(qmi_priv->uim_client, input, qmi_priv->context, &error);

    qmi_message_uim_logical_channel_input_unref(input);

    if (error)
    {
        strncpy(err, error->message, strlen(error->message));
        return -1;
    }

    if (!qmi_message_uim_logical_channel_output_get_result(output, &error))
    {
        strncpy(err, error->message, strlen(error->message));
        return -1;
    }

    /* Mark channel as having been cleaned up */
    if (channel == qmi_priv->last_channel_id)
    {
        qmi_priv->last_channel_id = 0;
    }

    qmi_message_uim_logical_channel_output_unref(output);

    return 0;
}
