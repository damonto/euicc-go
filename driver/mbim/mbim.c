// SPDX-License-Identifier: MIT
/*
 Copyright (c) 2024, Frans Klaver <frans.klaver@vislink.com>
 */

#include <libmbim-glib.h>
#include <stdio.h>
#include <stdint.h>

#include "mbim.h"

static void
async_result_ready(GObject *source_object,
                   GAsyncResult *res,
                   gpointer user_data)
{
    GAsyncResult **result_out = user_data;

    g_assert(*result_out == NULL);
    *result_out = g_object_ref(res);
}

static MbimDevice *
mbim_device_new_from_path(GFile *file,
                          GMainContext *context,
                          GError **error)
{
    g_autoptr(GMainContextPusher) pusher = NULL;
    g_autoptr(GAsyncResult) result = NULL;
    g_autofree gchar *id = NULL;

    pusher = g_main_context_pusher_new(context);

    id = g_file_get_path(file);
    if (id)
        mbim_device_new(file,
                        NULL,
                        async_result_ready,
                        &result);

    while (!result)
        g_main_context_iteration(context, TRUE);

    return mbim_device_new_finish(result, error);
}

static gboolean
mbim_device_open_sync(MbimDevice *device,
                      MbimDeviceOpenFlags open_flags,
                      GMainContext *context,
                      GError **error)
{
    g_autoptr(GMainContextPusher) pusher = NULL;
    g_autoptr(GAsyncResult) result = NULL;

    pusher = g_main_context_pusher_new(context);

    mbim_device_open_full(device,
                          open_flags,
                          15,
                          NULL,
                          async_result_ready,
                          &result);

    while (!result)
        g_main_context_iteration(context, TRUE);

    return mbim_device_open_finish(device, result, error);
}

static MbimMessage *
mbim_device_command_sync(MbimDevice *device, GMainContext *context, MbimMessage *request, GError **error)
{
    g_autoptr(GMainContextPusher) pusher = NULL;
    g_autoptr(GAsyncResult) result = NULL;

    pusher = g_main_context_pusher_new(context);

    mbim_device_command(device, request, 10, NULL, async_result_ready, &result);
    mbim_message_unref(request);

    while (result == NULL)
        g_main_context_iteration(context, TRUE);

    MbimMessage *response = mbim_device_command_finish(device, result, error);
    if (!response)
    {
        return NULL;
    }

    if (!mbim_message_response_get_result(response, MBIM_MESSAGE_TYPE_COMMAND_DONE, error))
    {
        return NULL;
    }

    return response;
}

static gboolean
mbim_device_close_sync(
    MbimDevice *device,
    GMainContext *context,
    GError **error)
{
    g_autoptr(GMainContextPusher) pusher = NULL;
    g_autoptr(GAsyncResult) result = NULL;

    pusher = g_main_context_pusher_new(context);

    mbim_device_close(device, 20, NULL, async_result_ready, &result);

    while (result == NULL)
        g_main_context_iteration(context, TRUE);

    return mbim_device_close_finish(device, result, error);
}

static gboolean is_sim_available(struct mbim_data *mbim_priv)
{
    MbimMessage *request = mbim_message_subscriber_ready_status_query_new(NULL);
    g_autoptr(MbimMessage) response = mbim_device_command_sync(
        mbim_priv->device, mbim_priv->context, request, NULL);
    if (!response)
        return FALSE;

    MbimSubscriberReadyState ready_state;
    if (!mbim_message_subscriber_ready_status_response_parse(
            response, &ready_state, NULL, NULL, NULL, NULL, NULL, NULL))
    {
        return FALSE;
    }

    switch (ready_state)
    {
    case MBIM_SUBSCRIBER_READY_STATE_NO_ESIM_PROFILE:
    case MBIM_SUBSCRIBER_READY_STATE_INITIALIZED:
        return TRUE;
    default:
        return FALSE;
    }
}

static int select_sim_slot(struct mbim_data *mbim_priv, char **err)
{
    g_autoptr(GError) error = NULL;

    MbimMessage *current_slot_request =
        mbim_message_ms_basic_connect_extensions_device_slot_mappings_query_new(NULL);

    g_autoptr(MbimMessage) current_slot_response = mbim_device_command_sync(
        mbim_priv->device, mbim_priv->context, current_slot_request, &error);
    if (!current_slot_response)
    {
        *err = strdup(error->message);
        return -1;
    }

    guint32 current_slot_count;
    g_autoptr(MbimSlotArray) current_slots = NULL;
    if (!mbim_message_ms_basic_connect_extensions_device_slot_mappings_response_parse(
            current_slot_response, &current_slot_count, &current_slots, &error))
    {
        *err = strdup(error->message);
        return -1;
    }

    if (current_slot_count && current_slots[0]->slot == mbim_priv->uim_slot)
    {
        return 0;
    }

    g_autoptr(GPtrArray) new_slot_array = g_ptr_array_new_with_free_func(g_free);
    MbimSlot *new_slot = g_new(MbimSlot, 1);
    new_slot->slot = mbim_priv->uim_slot;
    g_ptr_array_add(new_slot_array, new_slot);

    MbimMessage *update_slot_request = mbim_message_ms_basic_connect_extensions_device_slot_mappings_set_new(
        new_slot_array->len, (const MbimSlot **)new_slot_array->pdata, &error);
    if (!update_slot_request)
    {
        *err = strdup(error->message);
        return -1;
    }

    g_autoptr(MbimMessage) update_slot_response = mbim_device_command_sync(
        mbim_priv->device, mbim_priv->context, update_slot_request, &error);
    if (!update_slot_response)
    {
        *err = strdup(error->message);
        return -1;
    }

    guint32 slot_count;
    g_autoptr(MbimSlotArray) updated_slots = NULL;
    if (!mbim_message_ms_basic_connect_extensions_device_slot_mappings_response_parse(
            update_slot_response, &slot_count, &updated_slots, &error))
    {
        *err = strdup(error->message);
        return -1;
    }

    int retries = 20;
    while (retries--)
    {
        if (is_sim_available(mbim_priv))
        {
            return 0;
        }
        struct timespec ts = {
            .tv_sec = 0,
            .tv_nsec = 50000000};
        nanosleep(&ts, NULL);
    }

    *err = strdup("SIM not available");
    return -1;
}

int go_mbim_apdu_connect(struct mbim_data *mbim_priv, char *device_path, char **err)
{
    g_autoptr(GError) error = NULL;
    GFile *file;

    file = g_file_new_for_path(device_path);

    mbim_priv->context = g_main_context_new();

    mbim_priv->device = mbim_device_new_from_path(file, mbim_priv->context, &error);
    if (!mbim_priv->device)
    {
        *err = strdup(error->message);
        return -1;
    }

    MbimDeviceOpenFlags open_flags = MBIM_DEVICE_OPEN_FLAGS_NONE;
    if (mbim_priv->use_proxy)
        open_flags |= MBIM_DEVICE_OPEN_FLAGS_PROXY;

    mbim_device_open_sync(mbim_priv->device, open_flags, mbim_priv->context, &error);
    if (error)
    {
        *err = strdup(error->message);
        return -1;
    }

    return select_sim_slot(mbim_priv, err);
}

/*
 * Allocate storage in rx and copy the contents of response_data there. Also
 * tack the status at the end, as the MBIM protocol separates the status from
 * the rest of the response.
 */
static int copy_data_with_status(
    uint8_t **rx, uint32_t *rx_len,
    const guint8 *response_data, guint32 response_size,
    guint32 status)
{
    *rx_len = response_size + 2;
    *rx = malloc(*rx_len);
    if (!*rx)
        return -1;

    memcpy(*rx, response_data, response_size);
    (*rx)[*rx_len - 2] = status & 0xff;
    (*rx)[*rx_len - 1] = (status >> 8) & 0xff;

    return 0;
}

int go_mbim_apdu_transmit(struct mbim_data *mbim_priv, uint8_t **rx, uint32_t *rx_len, const uint8_t *tx, uint32_t tx_len, char **err)
{
    g_autoptr(GError) error = NULL;

    MbimMessage *request = mbim_message_ms_uicc_low_level_access_apdu_set_new(
        mbim_priv->last_channel_id,
        MBIM_UICC_SECURE_MESSAGING_NONE,
        MBIM_UICC_CLASS_BYTE_TYPE_INTER_INDUSTRY,
        tx_len,
        tx,
        &error);
    if (!request)
    {
        *err = strdup(error->message);
        return -1;
    }

    g_autoptr(MbimMessage) response = mbim_device_command_sync(
        mbim_priv->device, mbim_priv->context, request, &error);
    if (!response)
    {
        *err = strdup(error->message);
        return -1;
    }

    guint32 status = 0;
    guint32 response_size = 0;
    const guint8 *response_data = NULL;

    if (!mbim_message_ms_uicc_low_level_access_apdu_response_parse(
            response, &status, &response_size, &response_data, &error))
    {
        *err = strdup(error->message);
        return -1;
    }

    return copy_data_with_status(rx, rx_len, response_data, response_size, status);
}

int go_mbim_apdu_open_logical_channel(struct mbim_data *mbim_priv, const uint8_t *aid, uint8_t aid_len, char **err)
{
    g_autoptr(GError) error = NULL;
    guint8 channel_id;

    MbimMessage *request = mbim_message_ms_uicc_low_level_access_open_channel_set_new(
        aid_len, aid, 0, 1, &error);
    if (!request)
    {
        *err = strdup(error->message);
        return -1;
    }

    g_autoptr(MbimMessage) response = mbim_device_command_sync(
        mbim_priv->device, mbim_priv->context, request, &error);
    if (!response)
    {
        *err = strdup(error->message);
        return -1;
    }

    guint32 status = 0;
    guint32 channel = -1;
    guint32 response_size = 0;
    const guint8 *response_data = NULL;

    if (!mbim_message_ms_uicc_low_level_access_open_channel_response_parse(
            response, &status, &channel, &response_size, &response_data, &error))
    {
        *err = strdup(error->message);
        return -1;
    }

    mbim_priv->last_channel_id = channel;
    return channel;
}

int go_mbim_apdu_close_logical_channel(struct mbim_data *mbim_priv, uint8_t channel, char **err)
{
    g_autoptr(GError) error = NULL;

    MbimMessage *request = mbim_message_ms_uicc_low_level_access_close_channel_set_new(
        channel, 1, &error);
    if (!request)
    {
        *err = strdup(error->message);
        return -1;
    }

    g_autoptr(MbimMessage) response = mbim_device_command_sync(
        mbim_priv->device, mbim_priv->context, request, &error);
    if (!response)
    {
        *err = strdup(error->message);
        return -1;
    }

    guint32 status = 0;

    if (!mbim_message_ms_uicc_low_level_access_close_channel_response_parse(
            response, &status, &error))
    {
        *err = strdup(error->message);
        return -1;
    }

    if (channel == mbim_priv->last_channel_id)
        mbim_priv->last_channel_id = -1;

    return 0;
}

int go_mbim_apdu_disconnect(struct mbim_data *mbim_priv, char **err)
{
    g_autoptr(GError) error = NULL;
    int ret = 0;

    if (mbim_priv->last_channel_id > 0)
    {
        go_mbim_apdu_close_logical_channel(mbim_priv, mbim_priv->last_channel_id, err);
        mbim_priv->last_channel_id = -1;
    }

    mbim_device_close_sync(mbim_priv->device, mbim_priv->context, &error);
    if (error)
    {
        ret = -1;
        *err = strdup(error->message);
    }

    g_main_context_unref(mbim_priv->context);
    mbim_priv->context = NULL;
    mbim_priv->uim_slot = 0;

    return ret;
}
