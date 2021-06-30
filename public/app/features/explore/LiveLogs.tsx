import React, { PureComponent } from 'react';
import { css, cx } from '@emotion/css';
import tinycolor from 'tinycolor2';

import { LogMessageAnsi, Themeable, withTheme, getLogRowStyles, Icon, Button } from '@grafana/ui';
import { GrafanaTheme, LogRowModel, TimeZone, dateTimeFormat, LogLevel } from '@grafana/data';

import { ElapsedTime } from './ElapsedTime';

const LevelMapper = {
  [LogLevel.unknown]: 0,
  [LogLevel.trace]: 1,
  [LogLevel.debug]: 2,
  [LogLevel.info]: 3,
  [LogLevel.error]: 4,
  [LogLevel.warning]: 5,
  [LogLevel.critical]: 6,
};

const getStyles = (theme: GrafanaTheme) => ({
  logsRowsLive: css`
    label: logs-rows-live;
    font-family: ${theme.typography.fontFamily.monospace};
    font-size: ${theme.typography.size.sm};
    display: flex;
    flex-flow: column nowrap;
    height: 60vh;
    overflow-y: scroll;
    :first-child {
      margin-top: auto !important;
    }
  `,
  logsRowFade: css`
    label: logs-row-fresh;
    color: ${theme.colors.text};
    background-color: ${tinycolor(theme.palette.blue95).setAlpha(0.25).toString()};
    animation: fade 1s ease-out 1s 1 normal forwards;
    @keyframes fade {
      from {
        background-color: ${tinycolor(theme.palette.blue95).setAlpha(0.25).toString()};
      }
      to {
        background-color: transparent;
      }
    }
  `,
  logsRowsIndicator: css`
    font-size: ${theme.typography.size.md};
    padding-top: ${theme.spacing.sm};
    display: flex;
    align-items: center;
  `,
  button: css`
    margin-right: ${theme.spacing.sm};
  `,
  fullWidth: css`
    width: 100%;
  `,
});

export interface Props extends Themeable {
  logRows?: LogRowModel[];
  timeZone: TimeZone;
  stopLive: () => void;
  onPause: () => void;
  onResume: () => void;
  isPaused: boolean;
}

interface State {
  logRowsToRender?: LogRowModel[];
  sonify: boolean;
}

class LiveLogs extends PureComponent<Props, State> {
  private liveEndDiv: HTMLDivElement | null = null;
  private scrollContainerRef = React.createRef<HTMLTableSectionElement>();
  // HACK will run out of memory
  private sonifiedLines: any = {};

  constructor(props: Props) {
    super(props);
    this.state = {
      logRowsToRender: props.logRows,
      sonify: false,
    };
  }

  static getDerivedStateFromProps(nextProps: Props, state: State) {
    if (!nextProps.isPaused) {
      return {
        // We update what we show only if not paused. We keep any background subscriptions running and keep updating
        // our state, but we do not show the updates, this allows us start again showing correct result after resuming
        // without creating a gap in the log results.
        logRowsToRender: nextProps.logRows,
      };
    } else {
      return null;
    }
  }

  componentDidUpdate() {
    if (this.state.sonify && !this.props.isPaused) {
      const { logRowsToRender = [] } = this.state;
      const linesToSonify = [];
      const range = [LevelMapper[LogLevel.debug], LevelMapper[LogLevel.critical]];
      for (const row of logRowsToRender) {
        if (!this.sonifiedLines[row.uid]) {
          if (LevelMapper[row.logLevel] > LevelMapper[LogLevel.info]) {
            linesToSonify.push([row.timeEpochMs, LevelMapper[row.logLevel]]);
          }
        }
        this.sonifiedLines[row.uid] = true;
      }
      console.log('sonify lines', linesToSonify, 'range', range);
    }
  }

  onClickSonify = () => {
    this.setState((state) => ({ sonify: !state.sonify }));
  };

  /**
   * Handle pausing when user scrolls up so that we stop resetting his position to the bottom when new row arrives.
   * We do not need to throttle it here much, adding new rows should be throttled/buffered itself in the query epics
   * and after you pause we remove the handler and add it after you manually resume, so this should not be fired often.
   */
  onScroll = (event: React.SyntheticEvent) => {
    const { isPaused, onPause } = this.props;
    const { scrollTop, clientHeight, scrollHeight } = event.currentTarget;
    const distanceFromBottom = scrollHeight - (scrollTop + clientHeight);
    if (distanceFromBottom >= 5 && !isPaused) {
      onPause();
    }
  };

  rowsToRender = () => {
    const { isPaused } = this.props;
    let { logRowsToRender: rowsToRender = [] } = this.state;
    if (!isPaused) {
      // A perf optimisation here. Show just 100 rows when streaming and full length when the streaming is paused.
      rowsToRender = rowsToRender.slice(-100);
    }
    return rowsToRender;
  };

  render() {
    const { theme, timeZone, onPause, onResume, isPaused } = this.props;
    const { sonify } = this.state;
    const styles = getStyles(theme);
    const { logsRow, logsRowLocalTime, logsRowMessage } = getLogRowStyles(theme);

    return (
      <div>
        <table className={styles.fullWidth}>
          <tbody
            onScroll={isPaused ? undefined : this.onScroll}
            className={cx(['logs-rows', styles.logsRowsLive])}
            ref={this.scrollContainerRef}
          >
            {this.rowsToRender().map((row: LogRowModel) => {
              return (
                <tr className={cx(logsRow, styles.logsRowFade)} key={row.uid}>
                  <td className={cx(logsRowLocalTime)}>{dateTimeFormat(row.timeEpochMs, { timeZone })}</td>
                  <td className={cx(logsRowMessage)}>{row.hasAnsi ? <LogMessageAnsi value={row.raw} /> : row.entry}</td>
                </tr>
              );
            })}
            <tr
              ref={(element) => {
                this.liveEndDiv = element;
                // This is triggered on every update so on every new row. It keeps the view scrolled at the bottom by
                // default.
                if (this.liveEndDiv && !isPaused) {
                  this.scrollContainerRef.current?.scrollTo(0, this.scrollContainerRef.current.scrollHeight);
                }
              }}
            />
          </tbody>
        </table>
        <div className={styles.logsRowsIndicator}>
          <Button variant="secondary" onClick={isPaused ? onResume : onPause} className={styles.button}>
            <Icon name={isPaused ? 'play' : 'pause'} />
            &nbsp;
            {isPaused ? 'Resume' : 'Pause'}
          </Button>
          <Button variant="secondary" onClick={this.props.stopLive} className={styles.button}>
            <Icon name="square-shape" size="lg" type="mono" />
            &nbsp; Exit live mode
          </Button>
          <Button variant="secondary" onClick={this.onClickSonify} className={styles.button}>
            <Icon name="bell" />
            &nbsp; {sonify ? 'Stop sound' : 'Sonify log level'}
          </Button>
          {isPaused || (
            <span>
              Last line received: <ElapsedTime resetKey={this.props.logRows} humanize={true} /> ago
            </span>
          )}
        </div>
      </div>
    );
  }
}

export const LiveLogsWithTheme = withTheme(LiveLogs);
