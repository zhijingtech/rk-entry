// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"embed"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// GlobalAppCtx global application context
	GlobalAppCtx = &appContext{
		startTime: time.Now(),
		entries: map[string]map[string]Entry{
			appInfoEntryType: {
				appInfoEntryName: appInfoEntryDefault(),
			},
		},
		embedFS:       map[string]map[string]*embed.FS{},
		appInfoEntry:  appInfoEntryDefault(),
		shutdownSig:   make(chan os.Signal),
		shutdownHooks: make(map[string]ShutdownHook),
		userValues:    make(map[string]interface{}),
	}

	builtinRegFuncList = []RegFunc{
		registerAppInfoEntryYAML,
		RegisterLoggerEntryYAML,
		RegisterEventEntryYAML,
		RegisterConfigEntryYAML,
		// RegisterCertEntryYAML,
	}
	pluginRegFuncList   = make([]RegFunc, 0)
	webFrameRegFuncList = make([]RegFunc, 0)
	userDefRegFuncList  = make([]RegFunc, 0)

	LoggerEntryNoop   = NewLoggerEntryNoop()
	LoggerEntryStdout = NewLoggerEntryStdout()
	EventEntryNoop    = NewEventEntryNoop()
	EventEntryStdout  = NewEventEntryStdout()
)

// ShutdownHook defines interface of shutdown hook
type ShutdownHook func()

type ReadinessCheck func(req *http.Request, resp http.ResponseWriter) bool
type LivenessCheck func(req *http.Request, resp http.ResponseWriter) bool

// Init global app context with bellow fields.
func init() {
	signal.Notify(GlobalAppCtx.shutdownSig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
}

// Application context which contains bellow fields.
// It is not recommended override this value since StartTime would be assigned to current time
// at beginning of go process in init() function.
type appContext struct {
	startTime      time.Time                       `json:"-" yaml:"-"`
	appInfoEntry   *appInfoEntry                   `json:"-" yaml:"-"`
	readinessCheck ReadinessCheck                  `json:"-" yaml:"-"`
	livenessCheck  LivenessCheck                   `json:"-" yaml:"-"`
	entries        map[string]map[string]Entry     `json:"-" yaml:"-"`
	embedFS        map[string]map[string]*embed.FS `json:"-" yaml:"-"`
	userValues     map[string]interface{}          `json:"-" yaml:"-"`
	shutdownSig    chan os.Signal                  `json:"-" yaml:"-"`
	shutdownHooks  map[string]ShutdownHook         `json:"-" yaml:"-"`
}

// RegisterPluginRegFunc register rk plugins registration function.
// Call this while you need provided Entry needs to be registered and bootstrapped before user defined Entries.
func RegisterPluginRegFunc(regFunc RegFunc) {
	if regFunc == nil {
		return
	}
	pluginRegFuncList = append(pluginRegFuncList, regFunc)
}

// RegisterWebFrameRegFunc register rk web framework registration function.
// Call this while you need provided Entry needs to be registered and bootstrapped before user defined Entries.
func RegisterWebFrameRegFunc(regFunc RegFunc) {
	if regFunc == nil {
		return
	}
	webFrameRegFuncList = append(webFrameRegFuncList, regFunc)
}

// RegisterUserEntryRegFunc register user defined registration function.
// Call this while you need provided Entry needs to be registered and bootstrapped before user defined Entries.
func RegisterUserEntryRegFunc(regFunc RegFunc) {
	if regFunc == nil {
		return
	}
	userDefRegFuncList = append(userDefRegFuncList, regFunc)
}

// ListPluginEntryRegFunc list plugin registration functions.
func ListPluginEntryRegFunc() []RegFunc {
	// make a copy of the list
	res := make([]RegFunc, 0)
	for i := range pluginRegFuncList {
		res = append(res, pluginRegFuncList[i])
	}

	return res
}

// ListWebFrameEntryRegFunc list web framework registration functions.
func ListWebFrameEntryRegFunc() []RegFunc {
	// make a copy of the list
	res := make([]RegFunc, 0)
	for i := range webFrameRegFuncList {
		res = append(res, webFrameRegFuncList[i])
	}

	return res
}

// ListUserEntryRegFunc list web framework registration functions.
func ListUserEntryRegFunc() []RegFunc {
	// make a copy of the list
	res := make([]RegFunc, 0)
	for i := range userDefRegFuncList {
		res = append(res, userDefRegFuncList[i])
	}

	return res
}

// BootstrapBuiltInEntryFromYAML register and bootstrap builtin entries first
func BootstrapBuiltInEntryFromYAML(raw []byte) {
	ctx := context.Background()

	for i := range builtinRegFuncList {
		entries := builtinRegFuncList[i](raw)
		for _, v := range entries {
			v.Bootstrap(ctx)
		}
	}
}

// BootstrapPluginEntryFromYAML register and bootstrap plugin entries first
func BootstrapPluginEntryFromYAML(raw []byte) {
	ctx := context.Background()

	for i := range pluginRegFuncList {
		entries := pluginRegFuncList[i](raw)
		for _, v := range entries {
			v.Bootstrap(ctx)
		}
	}
}

// BootstrapWebFrameEntryFromYAML register and bootstrap web framework entries first
func BootstrapWebFrameEntryFromYAML(raw []byte) {
	ctx := context.Background()

	for i := range webFrameRegFuncList {
		entries := webFrameRegFuncList[i](raw)
		for _, v := range entries {
			v.Bootstrap(ctx)
		}
	}
}

// BootstrapUserEntryFromYAML register and bootstrap builtin entries first
func BootstrapUserEntryFromYAML(raw []byte) {
	ctx := context.Background()

	for i := range userDefRegFuncList {
		entries := userDefRegFuncList[i](raw)
		for _, v := range entries {
			v.Bootstrap(ctx)
		}
	}
}

// AddEmbedFS add embed.FS based on name and type of Entry
func (ctx *appContext) AddEmbedFS(entryType, entryName string, fs *embed.FS) {
	if len(entryType) < 1 || len(entryName) < 1 || fs == nil {
		return
	}

	if _, ok := ctx.embedFS[entryType]; !ok {
		ctx.embedFS[entryType] = make(map[string]*embed.FS)
	}

	ctx.embedFS[entryType][entryName] = fs
}

// GetEmbedFS get embed.FS based on name and type of Entry
func (ctx *appContext) GetEmbedFS(entryType, entryName string) *embed.FS {
	if v, ok := ctx.embedFS[entryType]; !ok {
		return nil
	} else {
		return v[entryName]
	}
}

// SetReadinessCheck set readiness check function
func (ctx *appContext) SetReadinessCheck(f ReadinessCheck) {
	ctx.readinessCheck = f
}

// SetLivenessCheck set liveness check function
func (ctx *appContext) SetLivenessCheck(f LivenessCheck) {
	ctx.livenessCheck = f
}

// ********************************
// ****** User value related ******
// ********************************

// AddValue add value to GlobalAppCtx.
func (ctx *appContext) AddValue(key string, value interface{}) {
	ctx.userValues[key] = value
}

// GetValue returns value from GlobalAppCtx.
func (ctx *appContext) GetValue(key string) interface{} {
	return ctx.userValues[key]
}

// ListValues list values from GlobalAppCtx.
func (ctx *appContext) ListValues() map[string]interface{} {
	return ctx.userValues
}

// RemoveValue remove value from GlobalAppCtx.
func (ctx *appContext) RemoveValue(key string) {
	delete(ctx.userValues, key)
}

// ClearValues clear values from GlobalAppCtx.
func (ctx *appContext) ClearValues() {
	for k := range ctx.userValues {
		delete(ctx.userValues, k)
	}
}

// ************************************
// ****** App info Entry related ******
// ************************************

func (ctx *appContext) GetAppInfoEntry() *appInfoEntry {
	return ctx.appInfoEntry
}

func (ctx *appContext) GetConfigEntry(entryName string) *ConfigEntry {
	entries := ctx.entries[ConfigEntryType]

	if v, ok := entries[entryName]; ok {
		return v.(*ConfigEntry)
	}

	return nil
}

func (ctx *appContext) GetLoggerEntry(entryName string) *LoggerEntry {
	entries := ctx.entries[LoggerEntryType]

	if v, ok := entries[entryName]; ok {
		return v.(*LoggerEntry)
	}

	return nil
}

// GetLoggerEntryDefault returns LoggerEntry marked as default.
// Return logger with STDOUT if no LoggerEntry was marked as default
func (ctx *appContext) GetLoggerEntryDefault() *LoggerEntry {
	res := LoggerEntryStdout

	entries := ctx.entries[LoggerEntryType]

	for _, v := range entries {
		if v.(*LoggerEntry).IsDefault {
			res = v.(*LoggerEntry)
			break
		}
	}

	return res
}

func (ctx *appContext) GetEventEntry(entryName string) *EventEntry {
	entries := ctx.entries[EventEntryType]

	if v, ok := entries[entryName]; ok {
		return v.(*EventEntry)
	}

	return nil
}

// GetEventEntryDefault returns EventEntry marked as default.
// Return logger with STDOUT if no EventEntry was marked as default
func (ctx *appContext) GetEventEntryDefault() *EventEntry {
	res := EventEntryStdout

	entries := ctx.entries[EventEntryType]

	for _, v := range entries {
		if v.(*EventEntry).IsDefault {
			res = v.(*EventEntry)
			break
		}
	}

	return res
}

// func (ctx *appContext) GetCertEntry(entryName string) *CertEntry {
// 	entries := ctx.entries[CertEntryType]

// 	if v, ok := entries[entryName]; ok {
// 		return v.(*CertEntry)
// 	}

// 	return nil
// }

func (ctx *appContext) AddEntry(entry Entry) {
	if entry == nil {
		return
	}

	if v, ok := ctx.entries[entry.GetType()]; !ok {
		ctx.entries[entry.GetType()] = map[string]Entry{
			entry.GetName(): entry,
		}
	} else {
		v[entry.GetName()] = entry
	}
}

func (ctx *appContext) clearEntries() {
	ctx.entries = map[string]map[string]Entry{}
}

func (ctx *appContext) GetEntry(entryType, entryName string) Entry {
	if v, ok := ctx.entries[entryType]; ok {
		return v[entryName]
	}

	return nil
}

func (ctx *appContext) RemoveEntry(entry Entry) {
	if entry == nil {
		return
	}

	if v, ok := ctx.entries[entry.GetType()]; ok {
		delete(v, entry.GetName())
	}
}

func (ctx *appContext) RemoveEntryByType(entryType string) {
	delete(ctx.entries, entryType)
}

func (ctx *appContext) ListEntriesByType(entryType string) map[string]Entry {
	if v, ok := ctx.entries[entryType]; ok {
		return v
	}

	return map[string]Entry{}
}

func (ctx *appContext) ListEntries() map[string]map[string]Entry {
	return ctx.entries
}

// func (ctx *appContext) GetSignerJwtEntry(entryName string) SignerJwt {
// 	if v := ctx.GetEntry(SignerJwtEntryType, entryName); v != nil {
// 		if res, ok := v.(SignerJwt); ok {
// 			return res
// 		}
// 	}

// 	return nil
// }

func (ctx *appContext) GetCryptoEntry(entryName string) Crypto {
	if v := ctx.GetEntry(CryptoEntryType, entryName); v != nil {
		if res, ok := v.(Crypto); ok {
			return res
		}
	}

	return nil
}

// ***********************************
// ****** Shutdown hook related ******
// ***********************************

// GetUpTime returns uptime of application from StartTime.
func (ctx *appContext) GetUpTime() time.Duration {
	return time.Since(ctx.startTime)
}

// GetStartTime returns start time of application.
func (ctx *appContext) GetStartTime() time.Time {
	return ctx.startTime
}

// AddShutdownHook add shutdown hook with name.
func (ctx *appContext) AddShutdownHook(name string, f ShutdownHook) {
	if f == nil {
		return
	}
	ctx.shutdownHooks[name] = f
}

// GetShutdownHook returns shutdown hook with name.
func (ctx *appContext) GetShutdownHook(name string) ShutdownHook {
	return ctx.shutdownHooks[name]
}

// ListShutdownHooks list shutdown hooks.
func (ctx *appContext) ListShutdownHooks() map[string]ShutdownHook {
	return ctx.shutdownHooks
}

// RemoveShutdownHook remove shutdown hook.
func (ctx *appContext) RemoveShutdownHook(name string) bool {
	if _, ok := GlobalAppCtx.shutdownHooks[name]; ok {
		delete(GlobalAppCtx.shutdownHooks, name)
		return true
	}

	return false
}

// Internal use only.
func (ctx *appContext) clearShutdownHooks() {
	for k := range ctx.shutdownHooks {
		delete(ctx.shutdownHooks, k)
	}
}

// *************************************
// ****** Shutdown sig related *********
// *************************************

// WaitForShutdownSig waits for shutdown signal.
func (ctx *appContext) WaitForShutdownSig() {
	<-ctx.shutdownSig
}

// GetShutdownSig returns shutdown signal.
func (ctx *appContext) GetShutdownSig() chan os.Signal {
	return ctx.shutdownSig
}
