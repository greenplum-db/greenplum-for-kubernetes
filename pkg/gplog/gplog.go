package gplog

import (
	"io"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func ForProd(debug bool) logr.Logger {
	return ctrlzap.New(func(o *ctrlzap.Options) {
		o.Development = debug
		o.Encoder = zapcore.NewJSONEncoder(newCustomEncoderConfig())
		o.ZapOpts = append(o.ZapOpts, zap.Hooks(zapHookFlushKlogOnFatal))
	})
}

func ForTest(logDest io.Writer) logr.Logger {
	return ctrlzap.New(func(o *ctrlzap.Options) {
		o.Development = true
		o.Encoder = zapcore.NewJSONEncoder(newCustomEncoderConfig())
		o.DestWritter = logDest
		o.ZapOpts = append(o.ZapOpts, zap.Hooks(zapHookFlushKlogOnFatal))
	})
}

func ForIntegration() logr.Logger {
	return ctrlzap.New(func(o *ctrlzap.Options) {
		o.Development = true
		o.Encoder = zapcore.NewConsoleEncoder(newCustomEncoderConfig())
		o.ZapOpts = append(o.ZapOpts, zap.Hooks(zapHookFlushKlogOnFatal))
	})
}

// klog periodically flushes its logs to files. If we log.Fatal() with logr/zap, then recent klog entries could be lost.
// This hook flushes klog whenever zap has a FatalLevel log to match klog's behaviour of flushing before exiting from
// a klog.Fatal().
func zapHookFlushKlogOnFatal(entry zapcore.Entry) error {
	if entry.Level == zap.FatalLevel {
		klog.Flush()
	}
	return nil
}

func newCustomEncoderConfig() zapcore.EncoderConfig {
	prodCfg := zap.NewProductionEncoderConfig()
	c := zap.NewDevelopmentEncoderConfig()
	c.TimeKey = prodCfg.TimeKey
	c.LevelKey = prodCfg.LevelKey
	c.NameKey = prodCfg.NameKey
	c.CallerKey = prodCfg.CallerKey
	c.MessageKey = prodCfg.MessageKey
	c.StacktraceKey = prodCfg.StacktraceKey
	return c
}
