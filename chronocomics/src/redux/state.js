/* @flow */

import type {Issue} from '../data';
export type {Issue} from '../data';

export type Filter = {
  +type: 'name';
  +query: string;
} | {
  +type: 'credit';
  +query: string;
} | {
  +type: 'character';
  +query: string;
}

export type AndFilter = $ReadOnlyArray<Filter>;
export type OrFilter = $ReadOnlyArray<AndFilter>;

export type State = {
  +filters: OrFilter;
  +results: $ReadOnlyArray<Issue>;
  +sorting: 'story' | 'year';
  +readIssues: $ReadOnlyArray<string>;
}

export function isIssueRead(issue: string, readIssues: $ReadOnlyArray<string>) {
  return readIssues.includes(issue);
}

export const defaultState: State = {
  filters: [[]],
  results: [],
  sorting: 'story',
  readIssues: [],
};
