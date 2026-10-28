package socks5

// PreservedSettingKeys identifies instance-local gateway configuration that must
// not be replaced by a backup from another instance. Credentials use the target
// instance encryption key, and imported listeners must not change its exposure.
func PreservedSettingKeys() []string {
	return []string{
		settingEnabled,
		settingListenAddress,
		settingPort,
		settingUsername,
		settingPassword,
		settingNodeID,
		settingSelection,
		settingRequireAuth,
		settingMaxAttempts,
		settingDialTimeoutSeconds,
		settingFailureCooldownSeconds,
		settingSpecificFallback,
		settingMaxConnections,
		settingMaxConnectionsPerClient,
		settingIdleTimeoutSeconds,
		settingMaxConnectionDurationSeconds,
		settingHealthCheckEnabled,
		settingHealthCheckIntervalSeconds,
		settingHealthCheckTimeoutSeconds,
		settingCandidateGroups,
		settingCandidateSources,
		settingCandidateProtocols,
		settingCandidateCountries,
		settingStickySessionEnabled,
		settingStickySessionMode,
		settingStickySessionTTLSeconds,
		settingAccounts,
		settingListeners,
		settingRoutingProfiles,
	}
}
