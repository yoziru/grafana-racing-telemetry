import {
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  isValidLiveChannelAddress,
  MetricFindValue,
  parseLiveChannelAddress,
  StreamingFrameOptions,
} from '@grafana/data';
import { DataSourceWithBackend, getGrafanaLiveSrv, getTemplateSrv } from '@grafana/runtime';

import { Observable, of, merge } from 'rxjs';

import { MyDataSourceJsonData, TelemetryQuery } from './types';

let counter = 100;

interface RecordingsResponse {
  recordings: string[];
}

export class DataSource extends DataSourceWithBackend<TelemetryQuery, MyDataSourceJsonData> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceJsonData>) {
    super(instanceSettings);
  }

  async metricFindQuery(_: TelemetryQuery): Promise<MetricFindValue[]> {
    // TODO: only handle specific queries (currently all queries)
    const response: RecordingsResponse = await this.getResource('recordings');
    return response.recordings.map((recordingId: string) => ({ text: recordingId }));
  }

  query(request: DataQueryRequest<TelemetryQuery>): Observable<DataQueryResponse> {
    const queries: Array<Observable<DataQueryResponse>> = [];
    for (const target of request.targets) {
      if (target.hide ?? false) {
        continue;
      }

      if (target.withStreaming ?? true) {
        // const templatedTarget = this.applyTemplateVariables(target)
        const { telemetry, graph } = target;
        const recording = getTemplateSrv().replace(target.recording);
        // if (!recording) {
        //   recording = getTemplateSrv().replace("$recording");
        // }
        // if (recording === "$recording") {
        //   recording= "live"
        // }
        const telemetryField = telemetry ?? 'Speed';

        const channel =
          recording.length > 0 && recording !== 'live'
            ? `ds/${this.uid}/${target.source}/${recording}`
            : `ds/${this.uid}/${target.source}`;
        const addr = parseLiveChannelAddress(channel);
        if (!isValidLiveChannelAddress(addr)) {
          continue;
        }

        // const maxLength = request.maxDataPoints ?? 500;
        // Reduce buffer size to improve performance on large dashboards
        const maxLength = graph ? request.maxDataPoints ?? 500 : 2;
        const buffer: StreamingFrameOptions = {
          maxDelta: request.range.to.valueOf() - request.range.from.valueOf(),
          maxLength,
        };

        const filter = {
          fields: ['time', telemetryField],
        };
        // if (telemetry === '*') {
        //   // for debugging purposes
        //   filter = null;
        // }

        queries.push(
          getGrafanaLiveSrv().getDataStream({
            key: `${request.requestId}.${counter++}`,
            addr: addr,
            filter,
            buffer,
          })
        );
      }
    }

    // With a single query just return the results
    if (queries.length === 1) {
      return queries[0];
    }
    if (queries.length > 1) {
      return merge(...queries);
    }
    return of(); // nothing
  }
}
