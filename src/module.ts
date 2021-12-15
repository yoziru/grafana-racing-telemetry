import { DataSourcePlugin } from '@grafana/data';

import { ConfigEditor } from './ConfigEditor';
import { DataSource } from './datasource';
import { QueryEditor } from './QueryEditor';
import { TelemetryQuery, MyDataSourceJsonData } from './types';

export const plugin = new DataSourcePlugin<DataSource, TelemetryQuery, MyDataSourceJsonData>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
