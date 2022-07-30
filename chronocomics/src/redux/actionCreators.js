/* @flow */
import type {Action} from './action';
import type {Filter} from './state';

export function setRead(link: string): Action {
  return {
    type: 'SET_READ',
    link,
  };
}

export function setUnread(link: string): Action {
  return {
    type: 'SET_UNREAD',
    link,
  };
}

export function removeFilter(i: number, j: number): Action {
  return {
    type: 'REMOVE_FILTER',
    i,
    j,
  };
}

export function addFilter(filter: Filter, andFilter: number): Action {
  return {
    type: 'ADD_FILTER',
    andFilter,
    filter,
  };
}

export function filterDone(): Action {
  return {
    type: 'FILTER_DONE',
  };
}

