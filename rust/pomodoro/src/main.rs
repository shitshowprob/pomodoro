use color_eyre::eyre::Result;
use ratatui::{
    DefaultTerminal, Frame,
    crossterm::event::{self, Event, KeyCode, KeyEvent},
    layout::{Alignment, Constraint, Layout},
    style::{Color, Style},
    text::Text,
    widgets::{Block, BorderType, List, ListState, Widget},
};
use std::{
    fmt::{self, Formatter},
    time::{self, Duration},
};

#[derive(Debug)]
struct App {
    state: AppState,
    exit: bool,
    duration: time::Duration,
    break_duration: time::Duration,
    items: Vec<DurationItem>,
    list_state: ListState,
    completed_sessions: u32,
    duration_type: DurationType,
}

#[derive(Debug)]
struct DurationItem {
    work_duration: time::Duration,
    break_duration: time::Duration,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum DurationType {
    Idle,
    Work,
    Break,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum AppState {
    Selecting,
    Running,
    Paused,
}

impl Default for DurationItem {
    fn default() -> Self {
        DurationItem {
            work_duration: Duration::from_secs(25 * 60),
            break_duration: Duration::from_secs(5 * 60),
        }
    }
}

impl fmt::Display for DurationItem {
    fn fmt(&self, f: &mut Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "{}min/{}min",
            self.work_duration.as_secs() / 60,
            self.break_duration.as_secs() / 60
        )
    }
}

impl App {
    fn new() -> Self {
        let items = vec![
            DurationItem::default(),
            DurationItem {
                work_duration: Duration::from_secs(5),
                break_duration: Duration::from_secs(3),
            },
        ];
        let mut list_state: ListState = ListState::default();
        list_state.select(Some(0));
        App {
            state: AppState::Selecting,
            exit: false,
            duration: items[0].work_duration,
            break_duration: items[0].break_duration,
            items: items,
            list_state: list_state,
            completed_sessions: 0,
            duration_type: DurationType::Idle,
        }
    }
    fn reset(&mut self) {
        self.duration = self.items[0].work_duration;
        self.break_duration = self.items[0].break_duration;
    }
    fn run(mut self, terminal: &mut DefaultTerminal) -> Result<()> {
        while !self.exit {
            terminal.draw(|frame| self.render(frame))?;
            if event::poll(Duration::from_secs_f64(1.0 / 50.0))? {
                match event::read()? {
                    Event::Key(key) => self.handle_key_events(key),
                    _ => {}
                }
            }
            match self.state {
                AppState::Paused => {}
                _ => self.tick(),
            }
        }
        Ok(())
    }

    fn tick(&mut self) {
        if self.duration == Duration::ZERO && self.duration_type == DurationType::Work {
            self.duration_type = DurationType::Break;
            self.duration = self.break_duration;
        } else if self.duration == Duration::ZERO && self.duration_type == DurationType::Break {
            self.duration_type = DurationType::Idle;
            self.duration = Duration::ZERO;
            self.completed_sessions += 1;
            self.state = AppState::Selecting;
        }
        self.duration = self
            .duration
            .saturating_sub(Duration::from_secs_f64(1.0 / 50.0));
    }
    fn handle_key_events(&mut self, key: KeyEvent) {
        match self.state {
            AppState::Selecting => match key.code {
                KeyCode::Char('q') => self.exit = true,
                KeyCode::Enter => {
                    self.state = AppState::Running;
                    self.duration_type = DurationType::Work;
                }
                KeyCode::Up | KeyCode::Char('k') => {
                    self.select_prev();
                }
                KeyCode::Down | KeyCode::Char('j') => {
                    self.select_next();
                }
                _ => {}
            },
            _ => {}
        }
        match key.code {
            KeyCode::Char('q') => self.exit = true,
            KeyCode::Char('p') => {
                if self.state == AppState::Paused {
                    self.state = AppState::Running;
                } else {
                    self.state = AppState::Paused;
                }
            }
            KeyCode::Esc => {
                self.state = AppState::Selecting;
                self.reset();
            }
            _ => {}
        }
    }
    fn select_next(&mut self) {
        let next_index = match self.list_state.selected() {
            Some(index) => {
                if index >= self.items.len() - 1 {
                    0
                } else {
                    index + 1
                }
            }
            None => 0,
        };
        self.list_state.select(Some(next_index));
        self.duration = self.items[next_index].work_duration;
        self.break_duration = self.items[next_index].break_duration;
    }
    fn select_prev(&mut self) {
        let prev_index = match self.list_state.selected() {
            Some(index) => {
                if index == 0 {
                    self.items.len() - 1
                } else {
                    index - 1
                }
            }
            None => 0,
        };
        self.list_state.select(Some(prev_index));
        self.duration = self.items[prev_index].work_duration;
        self.break_duration = self.items[prev_index].break_duration;
    }

    fn render(&mut self, frame: &mut Frame) {
        let block = Block::default()
            .title("Pomodoro")
            .title_alignment(Alignment::Center);
        let inner_block = block.inner(frame.area());
        block
            .border_type(BorderType::Plain)
            .render(frame.area(), frame.buffer_mut());
        let [render_area, below_area] =
            Layout::vertical([Constraint::Percentage(100), Constraint::Min(10)]).areas(inner_block);

        match self.state {
            AppState::Selecting => {
                let list = List::new(self.items.iter().map(|item| item.to_string()))
                    .block(Block::default().title("Select Pomodoro session"))
                    .highlight_style(Style::default().fg(Color::LightMagenta))
                    .highlight_symbol("> ");
                frame.render_stateful_widget(list, render_area, &mut self.list_state);
                if self.completed_sessions > 0 {
                    frame.render_widget(
                        Text::from(format!("Session completed: {}", self.completed_sessions)),
                        below_area,
                    );
                }
            }
            AppState::Running | AppState::Paused => {
                let milliseconds = self.duration.as_millis() % 1000;
                let seconds = self.duration.as_secs() % 60;
                let minutes = (self.duration.as_secs() / 60) % 60;
                let hours = (self.duration.as_secs() / 60) / 60;
                let time_left = match (hours, minutes, seconds) {
                    (0, 0, s) => format!("{}s {:03}ms", s, milliseconds),
                    (0, m, s) => format!("{}m {:02}s {:03}ms", m, s, milliseconds),
                    _ => format!(
                        "{}h {:02}m {:02}s {:03}ms",
                        hours, minutes, seconds, milliseconds
                    ),
                };
                let msg: String;
                if self.state == AppState::Paused {
                    msg = format!("Paused the timer: {}", time_left);
                } else {
                    msg = format!("🍅 {}", time_left);
                }
                let time_text = Text::from(msg)
                    .left_aligned()
                    .style(Style::default().bold().red());
                frame.render_widget(time_text, render_area);
            }
        }
    }
}
fn main() -> Result<()> {
    color_eyre::install()?;
    let app = App::new();
    let mut terminal = ratatui::init();
    app.run(&mut terminal)?;
    Ok(())
}
