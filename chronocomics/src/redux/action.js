/* @flow */
import type {Filter} from './state';

export type Action = {
	+type: 'ADD_FILTER';
	+andFilter: number;
  +filter: Filter;
} | {
  +type: 'FILTER_DONE';
} | {
	+type: 'REMOVE_FILTER';
  +i: number;
  +j: number;
} | {
  +type: 'SET_READ';
  +link: string;
} | {
  +type: 'SET_UNREAD';
  +link: string;
} | {
  +type: 'SET_SORTING';
  +sorting: 'story' | 'year';
};

