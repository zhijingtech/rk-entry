// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"embed"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestGetSpecificEntry(t *testing.T) {
	defer GlobalAppCtx.clearEntries()

	bootStr := `
---
config:
  - name: ut-config
logger:
  - name: ut-logger
event:
  - name: ut-event
cert:
  - name: ut-cert
`
	raw := []byte(bootStr)
	RegisterConfigEntryYAML(raw)
	RegisterLoggerEntryYAML(raw)
	RegisterEventEntryYAML(raw)
	// RegisterCertEntryYAML(raw)

	assert.NotNil(t, GlobalAppCtx.GetConfigEntry("ut-config"))
	assert.Nil(t, GlobalAppCtx.GetConfigEntry("ut-config-1"))

	assert.NotNil(t, GlobalAppCtx.GetLoggerEntry("ut-logger"))
	assert.Nil(t, GlobalAppCtx.GetLoggerEntry("ut-logger-1"))

	assert.NotNil(t, GlobalAppCtx.GetEventEntry("ut-event"))
	assert.Nil(t, GlobalAppCtx.GetEventEntry("ut-event-1"))

	// assert.NotNil(t, GlobalAppCtx.GetCertEntry("ut-cert"))
	// assert.Nil(t, GlobalAppCtx.GetCertEntry("ut-cert-1"))
}

func TestAppContext_RemoveEntryByType(t *testing.T) {
	defer GlobalAppCtx.clearEntries()

	bootStr := `
---
config:
  - name: ut-config
`
	raw := []byte(bootStr)
	RegisterConfigEntryYAML(raw)

	assert.Len(t, GlobalAppCtx.ListEntriesByType(ConfigEntryType), 1)
	GlobalAppCtx.RemoveEntryByType(ConfigEntryType)
	assert.Empty(t, GlobalAppCtx.ListEntriesByType(ConfigEntryType))
}

func TestGlobalAppCtx_init(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx)

	// validate start time recorded.
	assert.NotNil(t, GlobalAppCtx.GetStartTime())

	// validate appInfoEntry.
	assert.NotNil(t, GlobalAppCtx.GetAppInfoEntry())

	// validate builtin entry reg functions.
	assert.NotEmpty(t, builtinRegFuncList)

	// validate app info entry
	assert.NotNil(t, GlobalAppCtx.GetAppInfoEntry())

	// validate config entries.
	configEntries := GlobalAppCtx.ListEntriesByType(appInfoEntryType)
	assert.Equal(t, 0, len(configEntries))

	// validate zap logger entries.
	zapEntries := GlobalAppCtx.ListEntriesByType("non-exist")
	assert.Equal(t, 0, len(zapEntries))

	// validate shutdown hooks.
	assert.Empty(t, GlobalAppCtx.ListShutdownHooks())

	// validate user values.
	values := GlobalAppCtx.ListValues()
	assert.Equal(t, 0, len(values))
}

func TestRegisterEntryRegFunc_WithNilInput(t *testing.T) {
	length := len(pluginRegFuncList)
	RegisterPluginRegFunc(nil)
	assert.Len(t, pluginRegFuncList, length)
}

func TestRegisterEntryRegFunc_HappyCase(t *testing.T) {
	regFunc := func([]byte) map[string]Entry {
		return make(map[string]Entry)
	}

	length := len(pluginRegFuncList)

	RegisterPluginRegFunc(regFunc)
	assert.Len(t, pluginRegFuncList, length+1)
	// clear reg functions
	pluginRegFuncList = pluginRegFuncList[:0]
}

func TestListEntryRegFunc_HappyCase(t *testing.T) {
	regFunc := func([]byte) map[string]Entry {
		return make(map[string]Entry)
	}

	RegisterPluginRegFunc(regFunc)
	assert.Len(t, ListPluginEntryRegFunc(), 1)
	// clear reg functions
	pluginRegFuncList = pluginRegFuncList[:0]
}

// value related
func TestAppContext_AddValue_WithEmptyKey(t *testing.T) {
	key := ""
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	assert.Equal(t, value, GlobalAppCtx.GetValue(key).(string))
	GlobalAppCtx.ClearValues()
}

func TestAppContext_AddValue_WithEmptyValue(t *testing.T) {
	key := "key"
	value := ""
	GlobalAppCtx.AddValue(key, value)
	assert.Equal(t, value, GlobalAppCtx.GetValue(key).(string))
	GlobalAppCtx.ClearValues()
}

func TestAppContext_AddValue_HappyCase(t *testing.T) {
	key := "key"
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	assert.Equal(t, value, GlobalAppCtx.GetValue(key).(string))
	GlobalAppCtx.ClearValues()
}

func TestAppContext_GetValue_WithEmptyKey(t *testing.T) {
	key := ""
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	assert.Equal(t, value, GlobalAppCtx.GetValue(key).(string))
	GlobalAppCtx.ClearValues()
}

func TestAppContext_GetValue_WithEmptyValue(t *testing.T) {
	key := "key"
	value := ""
	GlobalAppCtx.AddValue(key, value)
	assert.Equal(t, value, GlobalAppCtx.GetValue(key).(string))
	GlobalAppCtx.ClearValues()
}

func TestAppContext_GetValue_HappyCase(t *testing.T) {
	key := "key"
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	assert.Equal(t, value, GlobalAppCtx.GetValue(key).(string))
	GlobalAppCtx.ClearValues()
}

func TestAppContext_ListValues_WithEmptyKey(t *testing.T) {
	key := ""
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	assert.True(t, len(GlobalAppCtx.ListValues()) == 1)
	assert.Equal(t, value, GlobalAppCtx.ListValues()[key])
	GlobalAppCtx.ClearValues()
}

func TestAppContext_ListValues_WithEmptyValue(t *testing.T) {
	key := "key"
	value := ""
	GlobalAppCtx.AddValue(key, value)
	assert.True(t, len(GlobalAppCtx.ListValues()) == 1)
	assert.Equal(t, value, GlobalAppCtx.ListValues()[key])
	GlobalAppCtx.ClearValues()
}

func TestAppContext_ListValues_HappyCase(t *testing.T) {
	key := "key"
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	assert.True(t, len(GlobalAppCtx.ListValues()) == 1)
	assert.Equal(t, value, GlobalAppCtx.ListValues()[key])
	GlobalAppCtx.ClearValues()
}

func TestAppContext_RemoveValue_WithNonExistValue(t *testing.T) {
	key := "key"
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	GlobalAppCtx.RemoveValue("non-exist-value")
	assert.True(t, len(GlobalAppCtx.ListValues()) == 1)

	GlobalAppCtx.ClearValues()
}

func TestAppContext_RemoveValue_HappyCase(t *testing.T) {
	key := "key"
	value := "value"
	GlobalAppCtx.AddValue(key, value)
	GlobalAppCtx.RemoveValue(key)
	assert.Empty(t, GlobalAppCtx.ListValues())

	GlobalAppCtx.ClearValues()
}

func TestAppContext_ClearValues_HappyCase(t *testing.T) {
	key := "key"
	value := "value"
	GlobalAppCtx.AddValue(key, value)

	GlobalAppCtx.ClearValues()
	assert.Empty(t, GlobalAppCtx.ListValues())
}

// shutdown signal related
func TestAppContext_GetShutdownSig_HappyCase(t *testing.T) {
	assert.NotNil(t, GlobalAppCtx.GetShutdownSig())
}

// shutdown hook related
func TestAppContext_AddShutdownHook_WithEmptyName(t *testing.T) {
	name := ""
	f := func() {}
	GlobalAppCtx.AddShutdownHook(name, f)
	assert.Equal(t, 1, len(GlobalAppCtx.ListShutdownHooks()))
	assert.NotNil(t, GlobalAppCtx.GetShutdownHook(name))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

func TestAppContext_AddShutdownHook_WithNilFunc(t *testing.T) {
	name := ""
	GlobalAppCtx.AddShutdownHook(name, nil)
	assert.Equal(t, 0, len(GlobalAppCtx.ListShutdownHooks()))
	assert.Nil(t, GlobalAppCtx.GetShutdownHook(name))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

func TestAppContext_AddShutdownHook_HappyCase(t *testing.T) {
	name := "unit-test-hook"
	f := func() {}
	GlobalAppCtx.AddShutdownHook(name, f)
	assert.Equal(t, 1, len(GlobalAppCtx.ListShutdownHooks()))
	assert.NotNil(t, GlobalAppCtx.GetShutdownHook(name))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

func TestAppContext_GetShutdownHook_WithNonExistHooks(t *testing.T) {
	name := "non-exist"
	assert.Nil(t, GlobalAppCtx.GetShutdownHook(name))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

func TestAppContext_GetShutdownHook_HappyCase(t *testing.T) {
	name := "unit-test-hook"
	f := func() {}
	GlobalAppCtx.AddShutdownHook(name, f)
	assert.NotNil(t, GlobalAppCtx.GetShutdownHook(name))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

func TestAppContext_ListShutdownHooks_WithEmptyHooks(t *testing.T) {
	assert.Equal(t, 0, len(GlobalAppCtx.ListShutdownHooks()))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

func TestAppContext_ListShutdownHooks_HappyCase(t *testing.T) {
	name := "unit-test-hook"
	f := func() {}
	GlobalAppCtx.AddShutdownHook(name, f)
	assert.Equal(t, 1, len(GlobalAppCtx.ListShutdownHooks()))
	// clear shutdown hooks
	GlobalAppCtx.clearShutdownHooks()
}

// entry related
func TestAppContext_AddEntry_WithEmptyName(t *testing.T) {
	defer GlobalAppCtx.clearEntries()

	name := "unit-test-entry"
	entry := &EntryMock{
		Name: name,
	}
	GlobalAppCtx.AddEntry(entry)
	assert.Equal(t, 1, len(GlobalAppCtx.ListEntries()))
	assert.Equal(t, entry, GlobalAppCtx.GetEntry(entry.GetType(), entry.GetName()))
}

func TestAppContext_AddEntry_WithNilEntry(t *testing.T) {
	defer GlobalAppCtx.clearEntries()

	GlobalAppCtx.AddEntry(nil)
	assert.Equal(t, 0, len(GlobalAppCtx.ListEntries()))
	assert.Nil(t, GlobalAppCtx.GetEntry("type", "name"))
}

func TestAppContext_AddEntry_HappyCase(t *testing.T) {
	defer GlobalAppCtx.clearEntries()

	entry := &EntryMock{
		Name: "unit-test-entry",
	}
	GlobalAppCtx.AddEntry(entry)
	assert.Equal(t, 1, len(GlobalAppCtx.ListEntries()))
	assert.Equal(t, entry, GlobalAppCtx.GetEntry(entry.GetType(), entry.GetName()))
}

func TestAppContext_GetEntry_HappyCase(t *testing.T) {
	defer GlobalAppCtx.clearEntries()

	entry := &EntryMock{
		Name: "unit-test-entry",
	}
	GlobalAppCtx.AddEntry(entry)
	assert.Equal(t, entry, GlobalAppCtx.GetEntry(entry.GetType(), entry.GetName()))
}

func TestAppContext_ListEntries_HappyCase(t *testing.T) {
	defer GlobalAppCtx.clearEntries()

	entry := &EntryMock{
		Name: "unit-test-entry",
	}
	GlobalAppCtx.AddEntry(entry)
	assert.Equal(t, 1, len(GlobalAppCtx.ListEntries()))
}

func TestAppContext_RemoveEntry(t *testing.T) {
	defer GlobalAppCtx.clearEntries()

	entry := &EntryMock{
		Name: "unit-test-entry",
	}
	GlobalAppCtx.AddEntry(entry)
	assert.Equal(t, entry, GlobalAppCtx.GetEntry(entry.GetType(), entry.GetName()))
	GlobalAppCtx.RemoveEntry(entry)
}

func TestAppContext_RemoveShutdownHook(t *testing.T) {
	assert.False(t, GlobalAppCtx.RemoveShutdownHook("non-exist"))
	GlobalAppCtx.AddShutdownHook("ut-shutdownhook", func() {})
	assert.True(t, GlobalAppCtx.RemoveShutdownHook("ut-shutdownhook"))
}

func TestAppContext_WaitForShutdownSig(t *testing.T) {
	go func() {
		time.Sleep(1 * time.Second)
		GlobalAppCtx.shutdownSig <- syscall.SIGTERM
	}()

	GlobalAppCtx.WaitForShutdownSig()
}

func TestAppContext_AddEmbedFS(t *testing.T) {
	// invalid case
	GlobalAppCtx.AddEmbedFS("", "name", &embed.FS{})
	assert.Empty(t, GlobalAppCtx.embedFS)

	GlobalAppCtx.AddEmbedFS("type", "", &embed.FS{})
	assert.Empty(t, GlobalAppCtx.embedFS)

	GlobalAppCtx.AddEmbedFS("type", "name", nil)
	assert.Empty(t, GlobalAppCtx.embedFS)

	// happy case
	GlobalAppCtx.AddEmbedFS("type", "name", &embed.FS{})
	assert.NotEmpty(t, GlobalAppCtx.embedFS)
	assert.NotNil(t, GlobalAppCtx.GetEmbedFS("type", "name"))
}

func TestAppContext_SetReadinessCheck(t *testing.T) {
	GlobalAppCtx.SetReadinessCheck(func(req *http.Request, resp http.ResponseWriter) bool {
		return true
	})
	GlobalAppCtx.SetLivenessCheck(func(req *http.Request, resp http.ResponseWriter) bool {
		return true
	})

	assert.NotNil(t, GlobalAppCtx.readinessCheck)
	assert.NotNil(t, GlobalAppCtx.livenessCheck)
}

type EntryMock struct {
	Name string
}

func (entry *EntryMock) Bootstrap(context.Context) {}

func (entry *EntryMock) Interrupt(context.Context) {}

func (entry *EntryMock) GetName() string {
	return entry.Name
}

func (entry *EntryMock) GetType() string {
	return "mock"
}

func (entry *EntryMock) String() string {
	return ""
}

func (entry *EntryMock) GetDescription() string {
	return ""
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
