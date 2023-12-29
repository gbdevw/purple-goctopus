package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gbdevw/purple-goctopus/spot/rest/account"
	"github.com/gbdevw/purple-goctopus/spot/rest/common"
	"github.com/gbdevw/purple-goctopus/spot/rest/earn"
	"github.com/gbdevw/purple-goctopus/spot/rest/funding"
	"github.com/gbdevw/purple-goctopus/spot/rest/market"
	"github.com/gbdevw/purple-goctopus/spot/rest/tracing"
	"github.com/gbdevw/purple-goctopus/spot/rest/trading"
	"github.com/gbdevw/purple-goctopus/spot/rest/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// A decorator for KrakenSpotRESTClient which manages code instrumentation by using the
// OpenTelemetry framework.
type KrakenSpotRESTClientInstrumentationDecorator struct {
	// Decorated
	decorated KrakenSpotRESTClientIface
	// Tracer
	tracer trace.Tracer
}

// # Description
//
// Decorate the provided KrakenSpotRESTClientIface implementation. The function returns the decorator which
// manages tracing and code instrumentation of the decorated by using the OpenTelemetry framework.
//
// # Inputs
//
//   - decorated: The KrakenSpotRESTClientIface implentation to decorate. Must no be nil.
//   - tracerProvider: Tracer provider to use to get the tracer used by the decorator to instrument code. If nil, the global tracer provider will be used (can be a NoopTracerProvider).
//
// # Returns
//
// The decorator which decorates the provided KrakenSpotRESTClientIface implementation.
func InstrumentKrakenSpotRESTClient(decorated KrakenSpotRESTClientIface, tracerProvider trace.TracerProvider) KrakenSpotRESTClientIface {
	if decorated == nil {
		// Panic if decorated is nil
		panic("decorated cannot be nil")
	}
	if tracerProvider == nil {
		// Use the global tracer provider if the provided tracer provider is nil.
		// In case the global tracer provider is not configured, its default behavior is to return a NoopTracerProvider.
		tracerProvider = otel.GetTracerProvider()
	}
	// Return decorator
	return &KrakenSpotRESTClientInstrumentationDecorator{
		decorated: decorated,
		tracer:    tracerProvider.Tracer(tracing.PackageName, trace.WithInstrumentationVersion(tracing.PackageVersion)),
	}
}

// Trace GetServerTime execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetServerTime(ctx context.Context) (*market.GetServerTimeResponse, *http.Response, error) {
	// Start a span
	ctx, span := dec.tracer.Start(ctx, tracing.TracesNamespace+".get_server_time", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetServerTime(ctx)
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetSystemStatus execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetSystemStatus(ctx context.Context) (*market.GetSystemStatusResponse, *http.Response, error) {
	// Start a span
	ctx, span := dec.tracer.Start(ctx, tracing.TracesNamespace+".get_system_status", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetSystemStatus(ctx)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttr := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttr = append(respAttr, attribute.String("status", resp.Result.Status))
		}
		span.AddEvent(tracing.TracesNamespace+".get_system_status.response", trace.WithAttributes(
			respAttr...,
		))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetAssetInfo execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetAssetInfo(ctx context.Context, opts *market.GetAssetInfoRequestOptions) (*market.GetAssetInfoResponse, *http.Response, error) {
	// Build attributes which record request settings
	reqAttr := []attribute.KeyValue{}
	if opts != nil {
		if len(opts.Assets) > 0 {
			reqAttr = append(reqAttr, attribute.StringSlice("assets", opts.Assets))
		}
		if opts.AssetClass != "" {
			reqAttr = append(reqAttr, attribute.String("aclass", opts.AssetClass))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_asset_info",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttr...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetAssetInfo(ctx, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttr := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttr = append(respAttr, attribute.Int("count", len(resp.Result)))
		}
		span.AddEvent(tracing.TracesNamespace+".get_asset_info.response", trace.WithAttributes(
			respAttr...,
		))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetTradableAssetPairs execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetTradableAssetPairs(ctx context.Context, opts *market.GetTradableAssetPairsRequestOptions) (*market.GetTradableAssetPairsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{}
	if opts != nil {
		if len(opts.Pairs) > 0 {
			reqAttributes = append(reqAttributes, attribute.StringSlice("pairs", opts.Pairs))
		}
		if opts.Info != "" {
			reqAttributes = append(reqAttributes, attribute.String("info", opts.Info))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_tradable_asset_pairs",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetTradableAssetPairs(ctx, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		span.AddEvent(tracing.TracesNamespace+".get_tradable_asset_pairs.response", trace.WithAttributes(
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetTickerInformation execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetTickerInformation(ctx context.Context, opts *market.GetTickerInformationRequestOptions) (*market.GetTickerInformationResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{}
	if opts != nil {
		if len(opts.Pairs) > 0 {
			reqAttributes = append(reqAttributes, attribute.StringSlice("pairs", opts.Pairs))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_ticker_information",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetTickerInformation(ctx, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		span.AddEvent(tracing.TracesNamespace+".get_ticker_information.response", trace.WithAttributes(
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetOHLCData execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetOHLCData(ctx context.Context, params market.GetOHLCDataRequestParameters, opts *market.GetOHLCDataRequestOptions) (*market.GetOHLCDataResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.String("pair", params.Pair)}
	if opts != nil {
		if opts.Interval != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("interval", opts.Interval))
		}
		if opts.Since != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("since", opts.Since))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_ohlc_data",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetOHLCData(ctx, params, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("pair", resp.Result.PairId),
				attribute.Int("count", len(resp.Result.Data)),
				attribute.Int64("last", resp.Result.Last))
		}
		span.AddEvent(tracing.TracesNamespace+".get_ohlc_data.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetOrderBook execute
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetOrderBook(ctx context.Context, params market.GetOrderBookRequestParameters, opts *market.GetOrderBookRequestOptions) (*market.GetOrderBookResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.String("pair", params.Pair)}
	if opts != nil {
		if opts.Count != 0 {
			reqAttributes = append(reqAttributes, attribute.Int("count", opts.Count))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_order_book",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetOrderBook(ctx, params, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("pair", resp.Result.PairId),
				attribute.Int("asks.count", len(resp.Result.Asks)),
				attribute.Int("bids.count", len(resp.Result.Bids)))
		}
		span.AddEvent(tracing.TracesNamespace+".get_order_book.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetRecentTrades execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetRecentTrades(ctx context.Context, params market.GetRecentTradesRequestParameters, opts *market.GetRecentTradesRequestOptions) (*market.GetRecentTradesResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.String("pair", params.Pair)}
	if opts != nil {
		if opts.Count != 0 {
			reqAttributes = append(reqAttributes, attribute.Int("count", opts.Count))
		}
		if opts.Since != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("since", opts.Since))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_recent_trades",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetRecentTrades(ctx, params, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("pair", resp.Result.PairId),
				attribute.Int("count", len(resp.Result.Trades)),
				attribute.Int64("last", resp.Result.Last))
		}
		span.AddEvent(tracing.TracesNamespace+".get_recent_trades.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetRecentSpreads execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetRecentSpreads(ctx context.Context, params market.GetRecentSpreadsRequestParameters, opts *market.GetRecentSpreadsRequestOptions) (*market.GetRecentSpreadsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.String("pair", params.Pair)}
	if opts != nil {
		if opts.Since != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("since", opts.Since))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_recent_spreads",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetRecentSpreads(ctx, params, opts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("pair", resp.Result.PairId),
				attribute.Int("count", len(resp.Result.Spreads)),
				attribute.Int64("last", resp.Result.Last))
		}
		span.AddEvent(tracing.TracesNamespace+".get_recent_spreads.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetAccountBalance execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetAccountBalance(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*account.GetAccountBalanceResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_account_balance",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetAccountBalance(ctx, nonce, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_account_balance.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetExtendedBalance execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetExtendedBalance(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*account.GetExtendedBalanceResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_extended_balance",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetExtendedBalance(ctx, nonce, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_extended_balance.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetTradeBalance execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetTradeBalance(ctx context.Context, nonce int64, opts *account.GetTradeBalanceRequestOptions, secopts *common.SecurityOptions) (*account.GetTradeBalanceResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		if opts.Asset != "" {
			reqAttributes = append(reqAttributes, attribute.String("asset", opts.Asset))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_trade_balance",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetTradeBalance(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
		}
		span.AddEvent(tracing.TracesNamespace+".get_trade_balance.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetOpenOrders execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetOpenOrders(ctx context.Context, nonce int64, opts *account.GetOpenOrdersRequestOptions, secopts *common.SecurityOptions) (*account.GetOpenOrdersResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("trades", opts.Trades))
		if opts.UserReference != nil {
			reqAttributes = append(reqAttributes, attribute.Int64("userref", *opts.UserReference))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_open_orders",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetOpenOrders(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", len(resp.Result.Open)))
		}
		span.AddEvent(tracing.TracesNamespace+".get_open_orders.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetClosedOrders execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetClosedOrders(ctx context.Context, nonce int64, opts *account.GetClosedOrdersRequestOptions, secopts *common.SecurityOptions) (*account.GetClosedOrdersResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("trades", opts.Trades))
		reqAttributes = append(reqAttributes, attribute.Bool("consolidate_taker", opts.ConsolidateTaker))
		if opts.UserReference != nil {
			reqAttributes = append(reqAttributes, attribute.Int64("userref", *opts.UserReference))
		}
		if opts.Closetime != "" {
			reqAttributes = append(reqAttributes, attribute.String("closetime", opts.Closetime))
		}
		if opts.Start != "" {
			reqAttributes = append(reqAttributes, attribute.String("start", opts.Start))
		}
		if opts.End != "" {
			reqAttributes = append(reqAttributes, attribute.String("end", opts.End))
		}
		if opts.Offset != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("ofs", opts.Offset))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_closed_orders",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetClosedOrders(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", len(resp.Result.Closed)))
		}
		span.AddEvent(tracing.TracesNamespace+".get_closed_orders.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace QueryOrdersInfo execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) QueryOrdersInfo(ctx context.Context, nonce int64, params account.QueryOrdersInfoParameters, opts *account.QueryOrdersInfoRequestOptions, secopts *common.SecurityOptions) (*account.QueryOrdersInfoResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.StringSlice("txid", params.TxId),
	}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("trades", opts.Trades))

		if opts.UserReference != nil {
			reqAttributes = append(reqAttributes, attribute.Int64("userref", *opts.UserReference))
		}
		if opts.ConsolidateTaker != nil {
			reqAttributes = append(reqAttributes, attribute.Bool("consolidate_taker", *opts.ConsolidateTaker))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".query_orders_info",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.QueryOrdersInfo(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".query_orders_info.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetTradesHistory execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetTradesHistory(ctx context.Context, nonce int64, opts *account.GetTradesHistoryRequestOptions, secopts *common.SecurityOptions) (*account.GetTradesHistoryResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("trades", opts.Trades))
		reqAttributes = append(reqAttributes, attribute.Bool("consolidate_taker", opts.ConsolidateTaker))
		if opts.Start != "" {
			reqAttributes = append(reqAttributes, attribute.String("start", opts.Start))
		}
		if opts.End != "" {
			reqAttributes = append(reqAttributes, attribute.String("end", opts.End))
		}
		if opts.Offset != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("ofs", opts.Offset))
		}
		if opts.Type != "" {
			reqAttributes = append(reqAttributes, attribute.String("type", opts.Type))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_trades_history",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetTradesHistory(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", len(resp.Result.Trades)))
		}
		span.AddEvent(tracing.TracesNamespace+".get_trades_history.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace QueryTradesInfo execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) QueryTradesInfo(ctx context.Context, nonce int64, params account.QueryTradesRequestParameters, opts *account.QueryTradesRequestOptions, secopts *common.SecurityOptions) (*account.QueryTradesInfoResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.StringSlice("txid", params.TransactionIds),
	}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("trades", opts.Trades))
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".query_trades_info",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.QueryTradesInfo(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".query_trades_info.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetOpenPositions execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetOpenPositions(ctx context.Context, nonce int64, opts *account.GetOpenPositionsRequestOptions, secopts *common.SecurityOptions) (*account.GetOpenPositionsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("docalcs", opts.DoCalcs))
		if len(opts.TransactionIds) > 0 {
			reqAttributes = append(reqAttributes, attribute.StringSlice("txid", opts.TransactionIds))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_open_positions",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetOpenPositions(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_open_positions.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetLedgersInfo execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetLedgersInfo(ctx context.Context, nonce int64, opts *account.GetLedgersInfoRequestOptions, secopts *common.SecurityOptions) (*account.GetLedgersInfoResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("without_count", opts.WithoutCount))
		if len(opts.Assets) > 0 {
			reqAttributes = append(reqAttributes, attribute.StringSlice("asset", opts.Assets))
		}
		if opts.AssetClass != "" {
			reqAttributes = append(reqAttributes, attribute.String("aclass", opts.AssetClass))
		}
		if opts.End != "" {
			reqAttributes = append(reqAttributes, attribute.String("end", opts.End))
		}
		if opts.Start != "" {
			reqAttributes = append(reqAttributes, attribute.String("start", opts.Start))
		}
		if opts.Type != "" {
			reqAttributes = append(reqAttributes, attribute.String("type", opts.Type))
		}
		if opts.Offset != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("ofs", opts.Offset))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_ledgers_info",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetLedgersInfo(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", len(resp.Result.Ledgers)))
		}
		span.AddEvent(tracing.TracesNamespace+".get_ledgers_info.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace QueryLedgers execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) QueryLedgers(ctx context.Context, nonce int64, params account.QueryLedgersRequestParameters, opts *account.QueryLedgersRequestOptions, secopts *common.SecurityOptions) (*account.QueryLedgersResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.StringSlice("id", params.Id),
	}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("trades", opts.Trades))
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".query_ledgers",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.QueryLedgers(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".query_ledgers.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetTradeVolume execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetTradeVolume(ctx context.Context, nonce int64, opts *account.GetTradeVolumeRequestOptions, secopts *common.SecurityOptions) (*account.GetTradeVolumeResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.StringSlice("pair", opts.Pairs))
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_trade_volume",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetTradeVolume(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
		}
		span.AddEvent(tracing.TracesNamespace+".get_trade_volume.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace RequestExportReport execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) RequestExportReport(ctx context.Context, nonce int64, params account.RequestExportReportRequestParameters, opts *account.RequestExportReportRequestOptions, secopts *common.SecurityOptions) (*account.RequestExportReportResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("report", params.Report),
		attribute.String("description", params.Description),
	}
	if opts != nil {
		if len(opts.Fields) > 0 {
			reqAttributes = append(reqAttributes, attribute.StringSlice("fields", opts.Fields))
		}
		if opts.Format != "" {
			reqAttributes = append(reqAttributes, attribute.String("format", opts.Format))
		}
		if opts.EndTm != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("endtm", opts.EndTm))
		}
		if opts.StartTm != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("starttm", opts.StartTm))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".request_export_report",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.RequestExportReport(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.String("id", resp.Result.Id))
		}
		span.AddEvent(tracing.TracesNamespace+".request_export_report.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetExportReportStatus execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetExportReportStatus(ctx context.Context, nonce int64, params account.GetExportReportStatusRequestParameters, secopts *common.SecurityOptions) (*account.GetExportReportStatusResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("report", params.Report),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_export_report_status",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetExportReportStatus(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_export_report_status.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace RetrieveDataExport execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) RetrieveDataExport(ctx context.Context, nonce int64, params account.RetrieveDataExportParameters, secopts *common.SecurityOptions) (*account.RetrieveDataExportResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("id", params.Id),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".retrieve_data_export",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.RetrieveDataExport(ctx, nonce, params, secopts)
	// Trace error and set span status
	tracing.TraceErrorAndSetStatus(span, err)
	// Return results
	return resp, httpresp, err
}

// Trace DeleteExportReport execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) DeleteExportReport(ctx context.Context, nonce int64, params account.DeleteExportReportRequestParameters, secopts *common.SecurityOptions) (*account.DeleteExportReportResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("id", params.Id),
		attribute.String("type", params.Type),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".delete_export_report",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.DeleteExportReport(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.Bool("cancel", resp.Result.Cancel),
				attribute.Bool("cancel", resp.Result.Delete))
		}
		span.AddEvent(tracing.TracesNamespace+".delete_export_report.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace AddOrder execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) AddOrder(ctx context.Context, nonce int64, params trading.AddOrderRequestParameters, opts *trading.AddOrderRequestOptions, secopts *common.SecurityOptions) (*trading.AddOrderResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("pair", params.Pair),
		attribute.String("ordertype", params.Order.OrderType),
		attribute.String("type", params.Order.Type),
		attribute.String("volume", params.Order.Volume),
		attribute.Bool("reduce_only", params.Order.ReduceOnly),
	}
	if params.Order.DisplayedVolume != "" {
		reqAttributes = append(reqAttributes, attribute.String("displayvol", params.Order.DisplayedVolume))
	}
	if params.Order.Price != "" {
		reqAttributes = append(reqAttributes, attribute.String("price", params.Order.Price))
	}
	if params.Order.Price2 != "" {
		reqAttributes = append(reqAttributes, attribute.String("price2", params.Order.Price2))
	}
	if params.Order.Trigger != "" {
		reqAttributes = append(reqAttributes, attribute.String("trigger", params.Order.Trigger))
	}
	if params.Order.Leverage != "" {
		reqAttributes = append(reqAttributes, attribute.String("leverage", params.Order.Leverage))
	}
	if params.Order.StpType != "" {
		reqAttributes = append(reqAttributes, attribute.String("stptype", params.Order.StpType))
	}
	if params.Order.OrderFlags != "" {
		reqAttributes = append(reqAttributes, attribute.String("oflags", params.Order.OrderFlags))
	}
	if params.Order.TimeInForce != "" {
		reqAttributes = append(reqAttributes, attribute.String("timeinforce", params.Order.TimeInForce))
	}
	if params.Order.ScheduledStartTime != "" {
		reqAttributes = append(reqAttributes, attribute.String("starttm", params.Order.ScheduledStartTime))
	}
	if params.Order.ExpirationTime != "" {
		reqAttributes = append(reqAttributes, attribute.String("expiretm", params.Order.ExpirationTime))
	}
	if params.Order.Close.OrderType != "" {
		reqAttributes = append(reqAttributes, attribute.String("close[ordertype]", params.Order.Close.OrderType))
	}
	if params.Order.Close.Price != "" {
		reqAttributes = append(reqAttributes, attribute.String("close[price]", params.Order.Close.Price))
	}
	if params.Order.Close.Price2 != "" {
		reqAttributes = append(reqAttributes, attribute.String("close[price2]", params.Order.Close.Price2))
	}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("validate", opts.Validate))
		if !opts.Deadline.IsZero() {
			reqAttributes = append(reqAttributes, attribute.String("deadline", opts.Deadline.Format(time.RFC3339)))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".add_order",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.AddOrder(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.StringSlice("txid", resp.Result.TransactionIDs),
				attribute.String("description", resp.Result.Description.Order),
				attribute.String("close", resp.Result.Description.Close))
		}
		span.AddEvent(tracing.TracesNamespace+".add_order.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace AddOrderBatch execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) AddOrderBatch(ctx context.Context, nonce int64, params trading.AddOrderBatchRequestParameters, opts *trading.AddOrderBatchRequestOptions, secopts *common.SecurityOptions) (*trading.AddOrderBatchResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("pair", params.Pair),
	}
	for index, order := range params.Orders {
		reqAttributes = append(reqAttributes,
			attribute.String(fmt.Sprintf("orders[%d][%s]", index, "ordertype"), order.OrderType),
			attribute.String(fmt.Sprintf("orders[%d][%s]", index, "type"), order.Type),
			attribute.String(fmt.Sprintf("orders[%d][%s]", index, "volume"), order.Volume),
			attribute.Bool(fmt.Sprintf("orders[%d][%s]", index, "reduce_only"), order.ReduceOnly))
		if order.DisplayedVolume != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "displayvol"), order.DisplayedVolume))
		}
		if order.Price != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "price"), order.Price))
		}
		if order.Price2 != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "price2"), order.Price2))
		}
		if order.Trigger != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "trigger"), order.Trigger))
		}
		if order.Leverage != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "leverage"), order.Leverage))
		}
		if order.StpType != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "stptype"), order.StpType))
		}
		if order.OrderFlags != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "oflags"), order.OrderFlags))
		}
		if order.TimeInForce != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "timeinforce"), order.TimeInForce))
		}
		if order.ScheduledStartTime != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "starttm"), order.ScheduledStartTime))
		}
		if order.ExpirationTime != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s]", index, "expiretm"), order.ExpirationTime))
		}
		if order.Close.OrderType != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "ordertype"), order.Close.OrderType))
		}
		if order.Close.Price != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "price"), order.Close.Price))
		}
		if order.Close.Price2 != "" {
			reqAttributes = append(reqAttributes, attribute.String(fmt.Sprintf("orders[%d][%s][%s]", index, "close", "price2"), order.Close.Price2))
		}
	}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("validate", opts.Validate))
		if !opts.Deadline.IsZero() {
			reqAttributes = append(reqAttributes, attribute.String("deadline", opts.Deadline.Format(time.RFC3339)))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".add_order_batch",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.AddOrderBatch(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", len(resp.Result.Orders)))
			for index, order := range resp.Result.Orders {
				respAttributes = append(
					respAttributes,
					attribute.String(fmt.Sprintf("orders[%d][%s]", index, "txid"), order.Id),
					attribute.String(fmt.Sprintf("orders[%d][%s]", index, "description"), order.Description.Order),
					attribute.String(fmt.Sprintf("orders[%d][%s]", index, "close"), order.Description.Close),
				)
			}
		}
		span.AddEvent(tracing.TracesNamespace+".add_order_batch.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace EditOrder execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) EditOrder(ctx context.Context, nonce int64, params trading.EditOrderRequestParameters, opts *trading.EditOrderRequestOptions, secopts *common.SecurityOptions) (*trading.EditOrderResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("pair", params.Pair),
		attribute.String("txid", params.Id),
	}
	if opts != nil {
		if opts.NewUserReference != "" {
			reqAttributes = append(reqAttributes, attribute.String("userref", opts.NewUserReference))
		}
		if opts.NewVolume != "" {
			reqAttributes = append(reqAttributes, attribute.String("volume", opts.NewVolume))
		}
		if opts.NewDisplayedVolume != "" {
			reqAttributes = append(reqAttributes, attribute.String("displayvol", opts.NewDisplayedVolume))
		}
		if opts.Price != "" {
			reqAttributes = append(reqAttributes, attribute.String("price", opts.Price))
		}
		if opts.Price2 != "" {
			reqAttributes = append(reqAttributes, attribute.String("price2", opts.Price2))
		}
		if len(opts.OFlags) > 0 {
			reqAttributes = append(reqAttributes, attribute.StringSlice("oflags", opts.OFlags))
		}
		reqAttributes = append(reqAttributes, attribute.Bool("validate", opts.Validate))
		reqAttributes = append(reqAttributes, attribute.Bool("cancel_response", opts.CancelResponse))
		if !opts.Deadline.IsZero() {
			reqAttributes = append(reqAttributes, attribute.String("deadline", opts.Deadline.Format(time.RFC3339)))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".edit_order",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.EditOrder(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			if resp.Result.Description != nil {
				respAttributes = append(
					respAttributes,
					attribute.String("description", resp.Result.Description.Order),
					attribute.String("close", resp.Result.Description.Close))
			}
			if resp.Result.TransactionID != "" {
				respAttributes = append(respAttributes, attribute.String("txid", resp.Result.TransactionID))
			}
			if resp.Result.NewUserReference != nil {
				respAttributes = append(respAttributes, attribute.Int64("newuserref", *resp.Result.NewUserReference))
			}
			if resp.Result.OldUserReference != nil {
				respAttributes = append(respAttributes, attribute.Int64("txid", *resp.Result.OldUserReference))
			}
			if resp.Result.OrdersCancelled != 0 {
				respAttributes = append(respAttributes, attribute.Int("orders_cancelled", resp.Result.OrdersCancelled))
			}
			if resp.Result.OriginalTransactionID != "" {
				respAttributes = append(respAttributes, attribute.String("originaltxid", resp.Result.OriginalTransactionID))
			}
			if resp.Result.Status != "" {
				respAttributes = append(respAttributes, attribute.String("status", resp.Result.Status))
			}
			if resp.Result.Volume != "" {
				respAttributes = append(respAttributes, attribute.String("volume", resp.Result.Volume))
			}
			if resp.Result.Price != "" {
				respAttributes = append(respAttributes, attribute.String("price", resp.Result.Price))
			}
			if resp.Result.Price2 != "" {
				respAttributes = append(respAttributes, attribute.String("price2", resp.Result.Price2))
			}
			if resp.Result.ErrorMsg != "" {
				respAttributes = append(respAttributes, attribute.String("error_message", resp.Result.ErrorMsg))
			}
		}
		span.AddEvent(tracing.TracesNamespace+".edit_order.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace CancelOrder execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) CancelOrder(ctx context.Context, nonce int64, params trading.CancelOrderRequestParameters, secopts *common.SecurityOptions) (*trading.CancelOrderResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("txid", params.Id),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".cancel_order",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.CancelOrder(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", resp.Result.Count))
		}
		span.AddEvent(tracing.TracesNamespace+".cancel_order.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace CancelAllOrders execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) CancelAllOrders(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*trading.CancelAllOrdersResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".cancel_all_orders",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.CancelAllOrders(ctx, nonce, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", resp.Result.Count))
		}
		span.AddEvent(tracing.TracesNamespace+".cancel_all_orders.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace CancelAllOrdersAfterX execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) CancelAllOrdersAfterX(ctx context.Context, nonce int64, params trading.CancelAllOrdersAfterXRequestParameters, secopts *common.SecurityOptions) (*trading.CancelAllOrdersAfterXResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.Int64("timeout", params.Timeout),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".cancel_all_orders_after_x",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.CancelAllOrdersAfterX(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("currentTime", resp.Result.CurrentTime.Format(time.RFC3339)),
				attribute.String("triggerTime", resp.Result.TriggerTime.Format(time.RFC3339)))
		}
		span.AddEvent(tracing.TracesNamespace+".cancel_all_orders_after_x.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace CancelOrderBatch execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) CancelOrderBatch(ctx context.Context, nonce int64, params trading.CancelOrderBatchRequestParameters, secopts *common.SecurityOptions) (*trading.CancelOrderBatchResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.StringSlice("orders", params.OrderIds),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".cancel_order_batch",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.CancelOrderBatch(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Int("count", resp.Result.Count))
		}
		span.AddEvent(tracing.TracesNamespace+".cancel_order_batch.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetDepositMethods execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetDepositMethods(ctx context.Context, nonce int64, params funding.GetDepositMethodsRequestParameters, secopts *common.SecurityOptions) (*funding.GetDepositMethodsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("asset", params.Asset),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_deposit_methods",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetDepositMethods(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_deposit_methods.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetDepositAddresses execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetDepositAddresses(ctx context.Context, nonce int64, params funding.GetDepositAddressesRequestParameters, opts *funding.GetDepositAddressesRequestOptions, secopts *common.SecurityOptions) (*funding.GetDepositAddressesResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("asset", params.Asset),
		attribute.String("method", params.Method),
	}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("new", opts.New))
		if opts.Amount != "" {
			reqAttributes = append(reqAttributes, attribute.String("amount", opts.Amount))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_deposit_addresses",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetDepositAddresses(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_deposit_addresses.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetStatusOfRecentDeposits execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetStatusOfRecentDeposits(ctx context.Context, nonce int64, opts *funding.GetStatusOfRecentDepositsRequestOptions, secopts *common.SecurityOptions) (*funding.GetStatusOfRecentDepositsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		if opts.Asset != "" {
			reqAttributes = append(reqAttributes, attribute.String("asset", opts.Asset))
		}
		if opts.Cursor != "" && opts.Cursor != "false" {
			reqAttributes = append(reqAttributes, attribute.String("cursor", opts.Cursor))
		}
		if opts.End != "" {
			reqAttributes = append(reqAttributes, attribute.String("end", opts.End))
		}
		if opts.Start != "" {
			reqAttributes = append(reqAttributes, attribute.String("start", opts.Start))
		}
		if opts.Method != "" {
			reqAttributes = append(reqAttributes, attribute.String("method", opts.Method))
		}
		if opts.Limit != 0 {
			reqAttributes = append(reqAttributes, attribute.Int64("limit", opts.Limit))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_status_of_recent_deposits",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetStatusOfRecentDeposits(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("next_cursor", resp.Result.NextCursor),
				attribute.Int("count", len(resp.Result.Deposits)),
			)
		}
		span.AddEvent(tracing.TracesNamespace+".get_status_of_recent_deposits.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetWithdrawalMethods execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetWithdrawalMethods(ctx context.Context, nonce int64, opts *funding.GetWithdrawalMethodsRequestOptions, secopts *common.SecurityOptions) (*funding.GetWithdrawalMethodsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		if opts.Asset != "" {
			reqAttributes = append(reqAttributes, attribute.String("asset", opts.Asset))
		}
		if opts.Network != "" {
			reqAttributes = append(reqAttributes, attribute.String("network", opts.Network))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_withdrawal_methods",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetWithdrawalMethods(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_withdrawal_methods.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetWithdrawalAddresses execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetWithdrawalAddresses(ctx context.Context, nonce int64, opts *funding.GetWithdrawalAddressesRequestOptions, secopts *common.SecurityOptions) (*funding.GetWithdrawalAddressesResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("verified", opts.Verified))
		if opts.Asset != "" {
			reqAttributes = append(reqAttributes, attribute.String("asset", opts.Asset))
		}
		if opts.Key != "" {
			reqAttributes = append(reqAttributes, attribute.String("key", opts.Key))
		}
		if opts.Method != "" {
			reqAttributes = append(reqAttributes, attribute.String("method", opts.Method))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_withdrawal_addresses",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetWithdrawalAddresses(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_withdrawal_addresses.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetWithdrawalInformation execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetWithdrawalInformation(ctx context.Context, nonce int64, params funding.GetWithdrawalInformationRequestParameters, secopts *common.SecurityOptions) (*funding.GetWithdrawalInformationResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("asset", params.Asset),
		attribute.String("key", params.Key),
		attribute.String("amount", params.Amount),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_withdrawal_information",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetWithdrawalInformation(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		span.AddEvent(tracing.TracesNamespace+".get_withdrawal_information.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace WithdrawFunds execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) WithdrawFunds(ctx context.Context, nonce int64, params funding.WithdrawFundsRequestParameters, opts *funding.WithdrawFundsRequestOptions, secopts *common.SecurityOptions) (*funding.WithdrawFundsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("asset", params.Asset),
		attribute.String("key", params.Key),
		attribute.String("amount", params.Amount),
	}
	if opts != nil {
		if opts.Address != "" {
			reqAttributes = append(reqAttributes, attribute.String("address", opts.Address))
		}
		if opts.MaxFee != "" {
			reqAttributes = append(reqAttributes, attribute.String("max_fee", opts.MaxFee))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".withdraw_funds",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.WithdrawFunds(ctx, nonce, params, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.String("refid", resp.Result.ReferenceID))
		}
		span.AddEvent(tracing.TracesNamespace+".withdraw_funds.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetStatusOfRecentWithdrawals execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetStatusOfRecentWithdrawals(ctx context.Context, nonce int64, opts *funding.GetStatusOfRecentWithdrawalsRequestOptions, secopts *common.SecurityOptions) (*funding.GetStatusOfRecentWithdrawalsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		if opts.Asset != "" {
			reqAttributes = append(reqAttributes, attribute.String("asset", opts.Asset))
		}
		if opts.End != "" {
			reqAttributes = append(reqAttributes, attribute.String("end", opts.End))
		}
		if opts.Start != "" {
			reqAttributes = append(reqAttributes, attribute.String("start", opts.Start))
		}
		if opts.Method != "" {
			reqAttributes = append(reqAttributes, attribute.String("method", opts.Method))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_status_of_recent_withdrawals",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetStatusOfRecentWithdrawals(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Int("count", len(resp.Result)),
		}
		span.AddEvent(tracing.TracesNamespace+".get_status_of_recent_withdrawals.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace RequestWithdrawalCancellation execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) RequestWithdrawalCancellation(ctx context.Context, nonce int64, params funding.RequestWithdrawalCancellationRequestParameters, secopts *common.SecurityOptions) (*funding.RequestWithdrawalCancellationResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("asset", params.Asset),
		attribute.String("refid", params.ReferenceId),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".request_withdrawal_cancellation",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.RequestWithdrawalCancellation(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Bool("result", resp.Result),
		}
		span.AddEvent(tracing.TracesNamespace+".request_withdrawal_cancellation.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace RequestWalletTransfer execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) RequestWalletTransfer(ctx context.Context, nonce int64, params funding.RequestWalletTransferRequestParameters, secopts *common.SecurityOptions) (*funding.RequestWalletTransferResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("asset", params.Asset),
		attribute.String("from", params.From),
		attribute.String("to", params.To),
		attribute.String("amount", params.Amount),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".request_wallet_transfer",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.RequestWalletTransfer(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.String("refid", resp.Result.ReferenceID))
		}
		span.AddEvent(tracing.TracesNamespace+".request_wallet_transfer.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace AllocateEarnFunds execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) AllocateEarnFunds(ctx context.Context, nonce int64, params earn.AllocateFundsRequestParameters, secopts *common.SecurityOptions) (*earn.AllocateFundsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("amount", params.Amount),
		attribute.String("strategy_id", params.StrategyId),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".allocate_earn_funds",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.AllocateEarnFunds(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Bool("result", resp.Result),
		}
		span.AddEvent(tracing.TracesNamespace+".allocate_earn_funds.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace DeallocateEarnFunds execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) DeallocateEarnFunds(ctx context.Context, nonce int64, params earn.DeallocateFundsRequestParameters, secopts *common.SecurityOptions) (*earn.DeallocateFundsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("amount", params.Amount),
		attribute.String("strategy_id", params.StrategyId),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".deallocate_earn_funds",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.DeallocateEarnFunds(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{
			attribute.StringSlice("error", resp.Error),
			attribute.Bool("result", resp.Result),
		}
		span.AddEvent(tracing.TracesNamespace+".deallocate_earn_funds.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetAllocationStatus execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetAllocationStatus(ctx context.Context, nonce int64, params earn.GetAllocationStatusRequestParameters, secopts *common.SecurityOptions) (*earn.GetAllocationStatusResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("strategy_id", params.StrategyId),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_allocation_status",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetAllocationStatus(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Bool("pending", resp.Result.Pending))
		}
		span.AddEvent(tracing.TracesNamespace+".get_allocation_status.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetDeallocationStatus execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetDeallocationStatus(ctx context.Context, nonce int64, params earn.GetDeallocationStatusRequestParameters, secopts *common.SecurityOptions) (*earn.GetDeallocationStatusResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{
		attribute.Int64("nonce", nonce),
		attribute.String("strategy_id", params.StrategyId),
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_deallocation_status",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetDeallocationStatus(ctx, nonce, params, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(respAttributes, attribute.Bool("pending", resp.Result.Pending))
		}
		span.AddEvent(tracing.TracesNamespace+".get_deallocation_status.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace ListEarnStrategies execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) ListEarnStrategies(ctx context.Context, nonce int64, opts *earn.ListEarnStrategiesRequestOptions, secopts *common.SecurityOptions) (*earn.ListEarnStrategiesResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes, attribute.Bool("ascending", opts.Ascending))
		if opts.Asset != "" {
			reqAttributes = append(reqAttributes, attribute.String("asset", opts.Asset))
		}
		if opts.Cursor != "" && opts.Cursor != "false" {
			reqAttributes = append(reqAttributes, attribute.String("cursor", opts.Cursor))
		}
		if opts.Limit != 0 {
			reqAttributes = append(reqAttributes, attribute.Int("limit", opts.Limit))
		}
		if opts.LockType != "" {
			reqAttributes = append(reqAttributes, attribute.String("lock_type", opts.LockType))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".list_earn_strategies",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.ListEarnStrategies(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("next_cursor", resp.Result.NextCursor),
				attribute.Int("count", len(resp.Result.Items)))
		}
		span.AddEvent(tracing.TracesNamespace+".list_earn_strategies.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace ListEarnAllocations execution.
func (dec *KrakenSpotRESTClientInstrumentationDecorator) ListEarnAllocations(ctx context.Context, nonce int64, opts *earn.ListEarnAllocationsRequestOptions, secopts *common.SecurityOptions) (*earn.ListEarnAllocationsResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	if opts != nil {
		reqAttributes = append(reqAttributes,
			attribute.Bool("ascending", opts.Ascending),
			attribute.Bool("hide_zero_allocations", opts.HideZeroAllocations),
		)
		if opts.ConvertedAsset != "" {
			reqAttributes = append(reqAttributes, attribute.String("converted_asset", opts.ConvertedAsset))
		}
	}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".list_earn_allocations",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.ListEarnAllocations(ctx, nonce, opts, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		if resp.Result != nil {
			respAttributes = append(
				respAttributes,
				attribute.String("converted_asset", resp.Result.ConvertedAsset),
				attribute.Int("count", len(resp.Result.Items)))
		}
		span.AddEvent(tracing.TracesNamespace+".list_earn_allocations.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}

// Trace GetWebsocketToken execution
func (dec *KrakenSpotRESTClientInstrumentationDecorator) GetWebsocketToken(ctx context.Context, nonce int64, secopts *common.SecurityOptions) (*websocket.GetWebsocketTokenResponse, *http.Response, error) {
	// Build attributes that will be added to span and that will record request settings
	reqAttributes := []attribute.KeyValue{attribute.Int64("nonce", nonce)}
	// Start a span
	ctx, span := dec.tracer.Start(
		ctx,
		tracing.TracesNamespace+".get_websocket_token",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(reqAttributes...))
	defer span.End()
	// Call decorated
	resp, httpresp, err := dec.decorated.GetWebsocketToken(ctx, nonce, secopts)
	// Add custom event and interesting values for received API response if any
	if resp != nil {
		respAttributes := []attribute.KeyValue{attribute.StringSlice("error", resp.Error)}
		span.AddEvent(tracing.TracesNamespace+".list_earn_allocations.response", trace.WithAttributes(respAttributes...))
	}
	// Trace error and set span status
	tracing.TraceApiOperationAndSetStatus(span, &resp.KrakenSpotRESTResponse, httpresp, err)
	// Return results
	return resp, httpresp, err
}
