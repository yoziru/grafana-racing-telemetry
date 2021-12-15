import React, { ChangeEvent, FC } from 'react';

import { DataSourcePluginOptionsEditorProps, onUpdateDatasourceJsonDataOption } from '@grafana/data';
import { Input, InlineField, FieldSet } from '@grafana/ui';

import { MyDataSourceJsonData, defaultMyDataSourceJsonData, MyDataSourceSettings } from './types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceJsonData> {}

export interface DataSourceConfigProps<J = MyDataSourceJsonData> extends DataSourcePluginOptionsEditorProps<J> {
  defaultRecordingBasePath?: string;
  defaultRecordingBufferDataPoints?: number;
}

export const DataSourceConfig: FC<DataSourceConfigProps> = (props: DataSourceConfigProps) => {
  const {
    options,
    onOptionsChange,
    defaultRecordingBasePath = defaultMyDataSourceJsonData.recordingBasePath,
    defaultRecordingBufferDataPoints = defaultMyDataSourceJsonData.recordingBufferDataPoints,
  } = props;

  const onRecordingBufferDataPointsChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...options.jsonData,
        recordingBufferDataPoints: parseInt(event.target.value, 10),
      },
    });
  };

  return (
    <FieldSet label="Connection Details" data-testid="connection-config">
      <InlineField label="Recording Base Path" labelWidth={28} tooltip="Where to store recordings of telemetry data">
        <Input
          className="width-30"
          placeholder={defaultRecordingBasePath}
          value={options.jsonData.recordingBasePath ?? defaultRecordingBasePath}
          onChange={onUpdateDatasourceJsonDataOption(props, 'recordingBasePath')}
        />
      </InlineField>

      <InlineField
        label="Recording Buffer Data Points"
        labelWidth={28}
        tooltip="How many datapoints to store on each recording (rolling buffer, e.g. 3600 datapoints at 60 FPS results in a 1 minute recording)"
      >
        <Input
          className="width-30"
          type="number"
          placeholder={defaultRecordingBufferDataPoints?.toString()}
          value={options.jsonData.recordingBufferDataPoints ?? defaultRecordingBufferDataPoints}
          onChange={onRecordingBufferDataPointsChange}
        />
      </InlineField>
    </FieldSet>
  );
};

export const ConfigEditor: FC<Props> = (props: Props) => {
  const onOptionsChange = (options: MyDataSourceSettings) => {
    props.onOptionsChange(options);
  };

  return (
    <>
      <DataSourceConfig {...props} onOptionsChange={onOptionsChange} />
    </>
  );
};
