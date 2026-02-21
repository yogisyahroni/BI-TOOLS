import { WebTracerProvider } from "@opentelemetry/sdk-trace-web";
import { getWebAutoInstrumentations } from "@opentelemetry/auto-instrumentations-web";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";
import {
  ConsoleSpanExporter,
  SimpleSpanProcessor,
  BatchSpanProcessor,
  SpanProcessor,
} from "@opentelemetry/sdk-trace-base";
import { registerInstrumentations } from "@opentelemetry/instrumentation";
import { ZoneContextManager } from "@opentelemetry/context-zone";
import { resourceFromAttributes } from "@opentelemetry/resources";
import { SemanticResourceAttributes } from "@opentelemetry/semantic-conventions";

const serviceName = "insight-engine-frontend";

export const initializeTracing = () => {
  if (typeof window === "undefined") return;

  // Exporter
  // Only enable if explicitly configured
  if (process.env.NEXT_PUBLIC_ENABLE_TRACING !== "true") {
    // eslint-disable-next-line no-console
    console.log("Tracing disabled (NEXT_PUBLIC_ENABLE_TRACING != true)");
    return;
  }

  const exporter = new OTLPTraceExporter({
    url: "http://localhost:4318/v1/traces", // Jaeger OTLP HTTP endpoint
  });

  const spanProcessors: SpanProcessor[] = [new BatchSpanProcessor(exporter)];

  // Also log to console for dev debugging
  if (process.env.NODE_ENV === "development") {
    spanProcessors.push(new SimpleSpanProcessor(new ConsoleSpanExporter()));
  }

  const provider = new WebTracerProvider({
    resource: resourceFromAttributes({
      [SemanticResourceAttributes.SERVICE_NAME]: serviceName,
    }),
    spanProcessors,
  });

  provider.register({
    contextManager: new ZoneContextManager(),
  });

  registerInstrumentations({
    instrumentations: [
      getWebAutoInstrumentations({
        // Prevent too much noise in console during dev
        "@opentelemetry/instrumentation-fetch": {
          propagateTraceHeaderCorsUrls: [
            new RegExp(/http:\/\/localhost:8080\/.*/), // Propagate trace headers to backend
          ],
        },
      }),
    ],
  });

  // eslint-disable-next-line no-console
  console.log("OpenTelemetry Tracing Initialized (OTLP -> localhost:4318)");
};
