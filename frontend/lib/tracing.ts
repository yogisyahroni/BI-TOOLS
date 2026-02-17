import { WebTracerProvider } from "@opentelemetry/sdk-trace-web";
import { getWebAutoInstrumentations } from "@opentelemetry/auto-instrumentations-web";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";
import {
  ConsoleSpanExporter,
  SimpleSpanProcessor,
  BatchSpanProcessor,
} from "@opentelemetry/sdk-trace-base";
import { registerInstrumentations } from "@opentelemetry/instrumentation";
import { ZoneContextManager } from "@opentelemetry/context-zone";
import { Resource } from "@opentelemetry/resources";
import { SemanticResourceAttributes } from "@opentelemetry/semantic-conventions";

const serviceName = "insight-engine-frontend";

export const initializeTracing = () => {
  if (typeof window === "undefined") return;

  const provider = new WebTracerProvider({
    // @ts-expect-error Resource type mismatch workaround
    resource: new Resource({
      [SemanticResourceAttributes.SERVICE_NAME]: serviceName,
    }),
  });
  // Exporter
  const exporter = new OTLPTraceExporter({
    url: "http://localhost:4318/v1/traces", // Jaeger OTLP HTTP endpoint
  });

  // Use BatchSpanProcessor for better performance in production
  // @ts-expect-error addSpanProcessor missing in type definition
  provider.addSpanProcessor(new BatchSpanProcessor(exporter));

  // Also log to console for dev debugging
  if (process.env.NODE_ENV === "development") {
    // @ts-expect-error addSpanProcessor missing in type definition
    provider.addSpanProcessor(new SimpleSpanProcessor(new ConsoleSpanExporter()));
  }

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
