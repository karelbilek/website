/* @flow */

import React from 'react';
import type {AndFilter, Filter as FilterData} from '../redux/state';
import {Filter} from './Filter';

type StaticProps = {
    filters: AndFilter;
}

type Props = StaticProps & {
    removeFilter: (j: number) => void;

    addFilter: ((filter: FilterData) => void);
    done: ?(() => void);
}

class FilterAdd extends React.PureComponent {
  props: {onSubmit: (filter: FilterData) => void};
  state: {type: string; query: string} = {type: 'name', query: ''};

  handleTypeChange(event) {
    this.setState({type: event.target.value});
  }

  handleQueryChange(event) {
    this.setState({query: event.target.value});
  }

  handleSubmit(event) {
    event.preventDefault();
    const type = this.state.type;
    const query = this.state.query;
    function getFilter(): ?FilterData {
      if (type === 'name') {
        return {type: 'name', query};
      }
      if (type === 'credit') {
        return {type: 'credit', query};
      }
      if (type === 'character') {
        return {type: 'character', query};
      }
      return null;
    }

    const f: ?FilterData = getFilter();
    if (f != null) {
      this.props.onSubmit(f);
    }
    this.setState({type: 'name', query: ''});
  }

  render() {
    return <form
      onSubmit={ (event) => this.handleSubmit(event) }
    >
      <select
        value={ this.state.type }
        onChange={ (event) => this.handleTypeChange(event) }
      >
        <option
          value='name'
        >name</option>
        <option
          value='credit'
        >credit</option>
        <option
          value='character'
        >character</option>
      </select>
      <input
        type='text'
        value={ this.state.query }
        onChange={ (event) => this.handleQueryChange(event) }
      />
      <input
        type='submit'
        value='add'
      />
    </form>;
  }
}

export class FilterRow extends React.PureComponent {
  props: Props;

  render() {
    const addFilterProp = this.props.addFilter;
    const addFilterComp = <FilterAdd
      onSubmit={ (filter) => addFilterProp(filter) }
    >
    </FilterAdd>;

    const doneProp = this.props.done;
    const doneComp = ((this.props.filters.length && this.props.done != null))
       ? <div
         className='next'
         onClick={ doneProp }
       >
         Add OR
       </div>
      : null;

    return <div>
      {this.props.filters.map((filter, i) =>
        <Filter
          key={ filter.type + filter.query + i }
          filter={ filter }
          remove={ () => this.props.removeFilter(i) }
        >
        </Filter>
      )}

      {addFilterComp}

      {doneComp}
    </div>;
  }
}
