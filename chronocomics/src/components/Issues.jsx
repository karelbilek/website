/* @flow */

import React from 'react';
import {connect} from 'react-redux';

import type {Issue as IssueData, State} from '../redux/state';
import * as actionCreators from '../redux/actionCreators';

import {Issue} from './Issue';
import {isIssueRead} from '../redux/state';

type StaticProps = {
  +data: $ReadOnlyArray<IssueData>;
  +readIssues: $ReadOnlyArray<string>;
}

type Props = StaticProps & {
    setRead: (link: string) => void;
    setUnread: (link: string) => void;
}

export class Issues extends React.PureComponent {
  props: Props;

  togglRead(link: string, currentState: boolean) {
    if (currentState) {
      this.props.setUnread(link);
    } else {
      this.props.setRead(link);
    }
  }

  render() {
    return <div>
      {this.props.data.map(issueData => {
        const read = isIssueRead(issueData.link, this.props.readIssues);
        return <Issue
          read = { read }
          data = { issueData }
          key = { issueData.link }
          toggl = { () => this.togglRead(issueData.link, read) }
        >
        </Issue>;
      })}
    </div>;
  }
}

function mapStateToProps(state: State): StaticProps {
  return {data: state.results, readIssues: state.readIssues};
}

export const IssuesContainer = connect(mapStateToProps, actionCreators)(Issues);
