import { DataQuery, DataSourceJsonData, DataSourceSettings } from '@grafana/data';

export interface TelemetryQuery extends DataQuery {
  telemetry: string[];
  source: string;
  recording?: string;
  withStreaming: boolean;
  graph: boolean;
}

export const defaultQuery: Partial<TelemetryQuery> = {
  telemetry: [],
  source: 'forzaHorizon5',
  recording: 'live',
  withStreaming: true,
  graph: false,
};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceJsonData extends DataSourceJsonData {
  recordingBasePath?: string;
  recordingBufferDataPoints?: number;
}

export const defaultMyDataSourceJsonData: Partial<MyDataSourceJsonData> = {
  recordingBasePath: '/var/lib/grafana/simracing-telemetry-recordings',
  recordingBufferDataPoints: 3600,
};

export interface MyDataSourceSecureJsonData {}
export type MyDataSourceSettings = DataSourceSettings<MyDataSourceJsonData, MyDataSourceSecureJsonData>;
