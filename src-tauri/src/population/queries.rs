use rusqlite::{params, Connection};
use tauri::State;

use crate::ez::{
    fs::db_file_path,
    types::{EzError, SessionManager},
};

use super::types::{
    ActivityCoordPayload, PersonPayload, PersonPlanPayload, QueryPopulationBboxPayload,
};

#[tauri::command]
pub fn query_population_bbox(
    source_name: String,
    min_lng: f64,
    min_lat: f64,
    max_lng: f64,
    max_lat: f64,
    limit: Option<i64>,
    offset: Option<i64>,
    manager: State<'_, SessionManager>,
) -> Result<QueryPopulationBboxPayload, EzError> {
    let connection = open_population_connection(&manager, &source_name)?;
    let limit = limit.unwrap_or(50);
    let offset = offset.unwrap_or(0);

    let mut rows_stmt = connection
        .prepare(
            "WITH paged_persons AS (
                SELECT DISTINCT person_id
                FROM plan_activity_coords
                WHERE lng >= ?1 AND lng <= ?2 AND lat >= ?3 AND lat <= ?4
                ORDER BY person_id ASC
                LIMIT ?5 OFFSET ?6
             )
             SELECT c.person_id, c.plan_id, c.plan_index, c.activity_index, c.lng, c.lat
             FROM plan_activity_coords c
             INNER JOIN paged_persons p ON c.person_id = p.person_id
             WHERE c.lng >= ?1 AND c.lng <= ?2 AND c.lat >= ?3 AND c.lat <= ?4
             ORDER BY c.person_id ASC, c.plan_index ASC, c.activity_index ASC",
        )
        .map_err(sqlite_err)?;

    let rows = rows_stmt
        .query_map(
            params![min_lng, max_lng, min_lat, max_lat, limit, offset],
            |row| {
                Ok(ActivityCoordPayload {
                    person_id: row.get(0)?,
                    plan_id: row.get(1)?,
                    plan_index: row.get(2)?,
                    activity_index: row.get(3)?,
                    lng: row.get(4)?,
                    lat: row.get(5)?,
                })
            },
        )
        .map_err(sqlite_err)?
        .collect::<Result<Vec<_>, _>>()
        .map_err(sqlite_err)?;

    let total_rows = connection
        .query_row(
            "SELECT COUNT(*)
             FROM plan_activity_coords
             WHERE lng >= ?1 AND lng <= ?2 AND lat >= ?3 AND lat <= ?4",
            params![min_lng, max_lng, min_lat, max_lat],
            |row| row.get::<_, i64>(0),
        )
        .map_err(sqlite_err)?;

    let total_people = connection
        .query_row(
            "SELECT COUNT(DISTINCT person_id)
             FROM plan_activity_coords
             WHERE lng >= ?1 AND lng <= ?2 AND lat >= ?3 AND lat <= ?4",
            params![min_lng, max_lng, min_lat, max_lat],
            |row| row.get::<_, i64>(0),
        )
        .map_err(sqlite_err)?;

    Ok(QueryPopulationBboxPayload {
        rows,
        total_rows,
        total_people,
    })
}

#[tauri::command]
pub fn search_population(
    source_name: String,
    query: String,
    exact: bool,
    limit: Option<i64>,
    offset: Option<i64>,
    manager: State<'_, SessionManager>,
) -> Result<QueryPopulationBboxPayload, EzError> {
    let connection = open_population_connection(&manager, &source_name)?;
    let limit = limit.unwrap_or(50);
    let offset = offset.unwrap_or(0);
    let search_param = if exact { query.clone() } else { format!("%{}%", query) };
    let op = if exact { "=" } else { "LIKE" };

    let rows_sql = format!(
        "WITH paged_persons AS (
            SELECT DISTINCT person_id
            FROM plan_activity_coords
            WHERE person_id {0} ?1
            ORDER BY person_id ASC
            LIMIT ?2 OFFSET ?3
         )
         SELECT c.person_id, c.plan_id, c.plan_index, c.activity_index, c.lng, c.lat
         FROM plan_activity_coords c
         INNER JOIN paged_persons p ON c.person_id = p.person_id
         WHERE c.person_id {0} ?1
         ORDER BY c.person_id ASC, c.plan_index ASC, c.activity_index ASC",
        op
    );
    let mut rows_stmt = connection.prepare(&rows_sql).map_err(sqlite_err)?;
    let rows = rows_stmt
        .query_map(params![search_param, limit, offset], |row| {
            Ok(ActivityCoordPayload {
                person_id: row.get(0)?,
                plan_id: row.get(1)?,
                plan_index: row.get(2)?,
                activity_index: row.get(3)?,
                lng: row.get(4)?,
                lat: row.get(5)?,
            })
        })
        .map_err(sqlite_err)?
        .collect::<Result<Vec<_>, _>>()
        .map_err(sqlite_err)?;

    let total_rows_sql = format!(
        "SELECT COUNT(*) FROM plan_activity_coords WHERE person_id {} ?1",
        op
    );
    let total_rows = connection
        .query_row(&total_rows_sql, params![search_param], |row| row.get::<_, i64>(0))
        .map_err(sqlite_err)?;

    let total_people_sql = format!(
        "SELECT COUNT(DISTINCT person_id) FROM plan_activity_coords WHERE person_id {} ?1",
        op
    );
    let total_people = connection
        .query_row(&total_people_sql, params![search_param], |row| row.get::<_, i64>(0))
        .map_err(sqlite_err)?;

    Ok(QueryPopulationBboxPayload {
        rows,
        total_rows,
        total_people,
    })
}

#[tauri::command]
pub fn get_person(
    source_name: String,
    person_id: String,
    manager: State<'_, SessionManager>,
) -> Result<PersonPayload, EzError> {
    let connection = open_population_connection(&manager, &source_name)?;

    let attributes_blob = connection
        .query_row(
            "SELECT attributes_blob
             FROM person_attributes
             WHERE person_id = ?1",
            params![person_id],
            |row| row.get::<_, Option<String>>(0),
        )
        .map_err(|err| match err {
            rusqlite::Error::QueryReturnedNoRows => EzError::PopulationPersonNotFound {
                person_id: person_id.clone(),
            },
            other => sqlite_err(other),
        })?;

    let mut stmt = connection
        .prepare(
            "SELECT plan_id, plan_index, selected, plan_blob
             FROM person_plans
             WHERE person_id = ?1
             ORDER BY plan_index ASC",
        )
        .map_err(sqlite_err)?;

    let plans = stmt
        .query_map(params![person_id.clone()], |row| {
            Ok(PersonPlanPayload {
                plan_id: row.get(0)?,
                plan_index: row.get(1)?,
                selected: row.get(2)?,
                plan_blob: row.get(3)?,
            })
        })
        .map_err(sqlite_err)?
        .collect::<Result<Vec<_>, _>>()
        .map_err(sqlite_err)?;

    Ok(PersonPayload {
        person_id,
        attributes_blob,
        plans,
    })
}

fn open_population_connection(
    manager: &State<'_, SessionManager>,
    source_name: &str,
) -> Result<Connection, EzError> {
    let guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_ref().ok_or(EzError::NoActiveSession)?;
    let path = db_file_path(&session.work_dir, source_name);
    if !path.exists() {
        return Err(EzError::SourceNotFound {
            name: source_name.to_string(),
        });
    }

    Connection::open(path).map_err(|err| EzError::Io {
        message: format!("Failed to open population SQLite database: {err}"),
    })
}

fn sqlite_err(err: rusqlite::Error) -> EzError {
    EzError::Io {
        message: format!("Population query failed: {err}"),
    }
}
