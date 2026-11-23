export function buildShareUpdateData(formData, share, subscriptionId) {
  const data = { ...formData, subscription_id: subscriptionId };
  if (share) {
    data.id = share.id;
    if (data.token === share.token) delete data.token;
  }
  return data;
}

export function resolveSelectedShares(ids, shares, retainedShares = []) {
  const byId = new Map([...retainedShares, ...shares].map((share) => [share.id, share]));
  return ids.map((id) => byId.get(id)).filter(Boolean);
}

export function buildShareBatchUpdates(mode, data, ids, shares) {
  if (ids.length === 0) return [];
  if (mode === 'devices' && data.maxDevices !== undefined) {
    return [{ ids, updates: { max_devices: Number(data.maxDevices) } }];
  }
  if (mode === 'expire' && [0, 1, 2].includes(data.expireType)) {
    const updates = { expire_type: data.expireType };
    if (data.expireType === 1) updates.expire_days = data.expireDays;
    if (data.expireType === 2) updates.expire_at = data.expireAt;
    return [{ ids, updates }];
  }
  if (mode === 'enabled') {
    if (data.action === 'enable' || data.action === 'disable') {
      return [{ ids, updates: { enabled: data.action === 'enable' } }];
    }
    if (data.action === 'toggle') {
      return [true, false]
        .map((enabled) => ({
          ids: ids.filter((id) => shares.some((share) => share.id === id && !!share.enabled !== enabled)),
          updates: { enabled }
        }))
        .filter((batch) => batch.ids.length > 0);
    }
  }
  return [];
}

export async function runShareBatchUpdates(batches, updateShares, fallbackError) {
  const results = [];
  for (const batch of batches) {
    try {
      const response = await updateShares(batch.ids, batch.updates);
      results.push(...(response?.results || batch.ids.map((id) => ({ id, success: true }))));
    } catch (error) {
      const response = error.response?.data;
      results.push(
        ...(response?.results || batch.ids.map((id) => ({ id, success: false, error: response?.msg || error.message || fallbackError })))
      );
    }
  }
  return results;
}
