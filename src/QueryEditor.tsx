import React, { FC, SyntheticEvent } from 'react';

import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { InlineField, InlineSwitch, MultiSelect, Select } from '@grafana/ui';

import { accOptions } from './accOptions';
import { DataSource } from './datasource';
import { dirtRallyOptions } from './dirtRallyOptions';
import { forzaHorizonOptions } from './forzaHorizonOptions';
import { iRacingOptions } from './iRacingOptions';
import { defaultQuery, MyDataSourceJsonData, TelemetryQuery } from './types';

export const sourceOptions = [
  { label: 'DiRT Rally 2.0', value: 'dirtRally2' },
  { label: 'Forza Horizon 5', value: 'forzaHorizon5' },
  { label: 'Assetto Corsa Competizione', value: 'acc' },
  { label: 'iRacing', value: 'iRacing' },
];

type Props = QueryEditorProps<DataSource, TelemetryQuery, MyDataSourceJsonData>;

export const QueryEditor: FC<Props> = (props: Props) => {
  const query = {
    ...defaultQuery,
    ...props.query,
  };
  const { telemetry, source, recording, withStreaming, graph } = query;

  let options = dirtRallyOptions;
  if (source === 'acc') {
    options = accOptions;
  } else if (source === 'forzaHorizon5') {
    options = forzaHorizonOptions;
  } else if (source === 'iRacing') {
    options = iRacingOptions;
  }

  const recordings = ['live', '$recording'];
  const recordingOptions = recordings.map((o) => ({ label: o, value: o }));

  const onTelemetryChange = (option: Array<SelectableValue<string>>): void => {
    const { onChange, onRunQuery } = props;
    onChange({ ...query, telemetry: Array.from(option.map((o) => o.value).filter((o): o is string => Boolean(o))) });
    // executes the query
    onRunQuery();
  };

  const onSourceChange = (option: SelectableValue<string>): void => {
    const { onChange, onRunQuery } = props;
    onChange({ ...query, source: option.value ?? 'live' });
    onRunQuery();
  };

  const onRecordingChange = (option: SelectableValue<string>): void => {
    const { onChange, onRunQuery } = props;
    onChange({ ...query, recording: option.value });
    onRunQuery();
  };

  const onWithStreamingChange = (event: SyntheticEvent<HTMLInputElement>): void => {
    const { onChange, onRunQuery } = props;
    onChange({ ...query, withStreaming: event.currentTarget.checked });
    // executes the query
    onRunQuery();
  };

  const onGraphChange = (event: SyntheticEvent<HTMLInputElement>): void => {
    const { onChange, onRunQuery } = props;
    onChange({ ...query, graph: event.currentTarget.checked });
    // executes the query
    onRunQuery();
  };

  return (
    <>
      <div className="gf-form">
        <InlineField label="Fields">
          <MultiSelect options={options} value={telemetry} onChange={onTelemetryChange} defaultValue={[]} />
        </InlineField>
      </div>
      <div className="gf-form">
        <InlineField label="Source">
          <Select options={sourceOptions} value={source} onChange={onSourceChange} defaultValue={'acc'} />
        </InlineField>
        <InlineField label="From">
          <Select
            options={recordingOptions}
            value={recording}
            allowCustomValue={true}
            onChange={onRecordingChange}
            defaultValue={'live'}
          />
        </InlineField>
        <InlineField label="Enable streaming">
          <InlineSwitch value={withStreaming ?? false} onChange={onWithStreamingChange} />
        </InlineField>
        <InlineField label="Graph">
          <InlineSwitch value={graph} onChange={onGraphChange} />
        </InlineField>
      </div>
    </>
  );
};
