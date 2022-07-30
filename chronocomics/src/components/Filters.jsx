/* @flow */

import React from 'react';
import {connect} from 'react-redux';

import type {State, OrFilter, Filter as FilterData} from '../redux/state';
import * as actionCreators from '../redux/actionCreators';

import {FilterRow} from './FilterRow';

type StaticProps = {
  +filters: OrFilter;
}

type Props = StaticProps & {
    removeFilter: (i: number, j: number) => void;
    addFilter: (filter: FilterData) => void;
    filterDone: () => void;
}

export class Filters extends React.PureComponent {
  props: Props;

  render() {
    return <div>
      {this.props.filters.map((filters, i) => {
        const last = i === this.props.filters.length - 1;
        return <div
          key={ i.toString() }
        >
          <FilterRow
            filters={ filters }
            removeFilter={ (j) => this.props.removeFilter(i, j) }
            addFilter={ (filter) => this.props.addFilter(filter, i) }
            done={ last ? this.props.filterDone : null }
          ></FilterRow>
          {last ? '' : 'OR'}
        </div>;
      })}
    </div>;
  }
}

function mapStateToProps(state: State): StaticProps {
  return {filters: state.filters};
}

export const FiltersContainer = connect(mapStateToProps, actionCreators)(Filters);
