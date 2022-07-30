/* @flow */
import type {State} from './state';
import type {Action} from './action';
import {defaultState} from './state';
import {getData} from '../data';

function replace<X>(arr: $ReadOnlyArray<X>, i: number, fun: (inp: X) => X): $ReadOnlyArray<X> {
  return [...arr.slice(0, i), fun(arr[i]), ...arr.slice(i + 1)];
}

function remove<X>(arr: $ReadOnlyArray<X>, i: number): $ReadOnlyArray<X> {
  return [...arr.slice(0, i), ...arr.slice(i + 1)];
}

function remove2d<X>(arr: $ReadOnlyArray<$ReadOnlyArray<X>>, i: number, j: number): $ReadOnlyArray<$ReadOnlyArray<X>> {
  return replace(arr, i, subarr => remove(subarr, j));
}

function replaceI<X>(arr: $ReadOnlyArray<X>, i: number, fun: (inp: X) => X): $ReadOnlyArray<X> {
  return replace(arr, i, fun);
}

function addToI<X>(arr: $ReadOnlyArray<$ReadOnlyArray<X>>, i: number, elem: X): $ReadOnlyArray<$ReadOnlyArray<X>> {
  return replaceI(arr, i, subarr => [...subarr, elem]);
}

export function reducer(state: State = defaultState, action: Action): State {
  if (action.type === 'ADD_FILTER') {
    if (state.filters.length === 0) {
      throw new Error('No place to add new filter');
    }
    const filters = addToI(state.filters, action.andFilter, action.filter);
    const results = getData(filters, state.sorting);
    return {
      ...state,
      filters,
      results,
    };
  }

  if (action.type === 'FILTER_DONE') {
    return {
      ...state,
      filters: [...state.filters, []],
    };
  }

  if (action.type === 'REMOVE_FILTER') {
    const filtersMaybeEmpty = remove2d(state.filters, action.i, action.j).filter(f => f.length !== 0);
    const filters = filtersMaybeEmpty.length === 0 ? [[]] : filtersMaybeEmpty;
    const results = getData(filters, state.sorting);
    return {
      ...state,
      filters,
      results,
      sorting: state.sorting,
    };
  }
  if (action.type === 'SET_SORTING') {
    const results = getData(state.filters, action.sorting);
    return {
      ...state,
      results,
      sorting: action.sorting,
    };
  }

  // note - doesn't check double
  if (action.type === 'SET_READ') {
    return {
      ...state,
      readIssues: [...new Set(state.readIssues).add(action.link)],
    };
  }
  if (action.type === 'SET_UNREAD') {
    const readIssues = new Set(state.readIssues);
    readIssues.delete(action.link);
    return {
      ...state,
      readIssues: [...readIssues],
    };
  }
  return state;
}
