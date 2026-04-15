use rand::Rng;
use rusqlite::{params, Connection, Transaction};
use tauri::State;

use crate::ez::{
    fs::db_file_path,
    types::{EzError, SessionManager},
};

use crate::ez::current_crs;
use crate::projection::Projector;

use super::{
    patch::apply_plan_edits_to_blob,
    types::{ApplyPlanEditsInput, PersonPayload, PersonPlanPayload},
};

#[tauri::command]
pub fn update_person_attributes(
    source_name: String,
    person_id: String,
    new_blob: Option<String>,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let path = population_db_path(&manager, &source_name)?;
    let connection = Connection::open(path).map_err(io_err)?;

    let affected = connection
        .execute(
            "UPDATE person_attributes
             SET attributes_blob = ?1
             WHERE person_id = ?2",
            params![new_blob, person_id.clone()],
        )
        .map_err(sqlite_err)?;

    if affected == 0 {
        return Err(EzError::PopulationPersonNotFound { person_id });
    }

    mark_dirty(&manager)?;
    Ok(())
}

#[tauri::command]
pub fn apply_plan_edits(
    source_name: String,
    person_id: String,
    plan_id: String,
    edits: ApplyPlanEditsInput,
    manager: State<'_, SessionManager>,
) -> Result<PersonPlanPayload, EzError> {
    let path = population_db_path(&manager, &source_name)?;
    let mut connection = Connection::open(path).map_err(io_err)?;
    let tx = connection.transaction().map_err(sqlite_err)?;

    let (plan_index, selected, plan_blob) = tx
        .query_row(
            "SELECT plan_index, selected, plan_blob
             FROM person_plans
             WHERE person_id = ?1 AND plan_id = ?2",
            params![person_id.clone(), plan_id.clone()],
            |row| {
                Ok((
                    row.get::<_, i64>(0)?,
                    row.get::<_, i64>(1)?,
                    row.get::<_, String>(2)?,
                ))
            },
        )
        .map_err(|err| match err {
            rusqlite::Error::QueryReturnedNoRows => EzError::PopulationPlanNotFound {
                person_id: person_id.clone(),
                plan_id: plan_id.clone(),
            },
            other => sqlite_err(other),
        })?;

    let crs = current_crs(&manager)?;
    let projector = Projector::new(&crs)?;
    let new_blob = apply_plan_edits_to_blob(&plan_blob, &edits, &projector)?;

    tx.execute(
        "UPDATE person_plans
         SET plan_blob = ?1
         WHERE person_id = ?2 AND plan_id = ?3",
        params![new_blob, person_id.clone(), plan_id.clone()],
    )
    .map_err(sqlite_err)?;

    tx.execute(
        "DELETE FROM plan_activity_coords
         WHERE person_id = ?1 AND plan_id = ?2",
        params![person_id.clone(), plan_id.clone()],
    )
    .map_err(sqlite_err)?;

    {
        let mut stmt = tx
            .prepare(
                "INSERT INTO plan_activity_coords
                 (person_id, plan_id, plan_index, activity_index, lng, lat)
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
            )
            .map_err(sqlite_err)?;
        for (activity_index, activity) in edits.activities.iter().enumerate() {
            if let (Some(lng), Some(lat)) = (activity.lng, activity.lat) {
                let _ = projector.unproject_lng_lat(lng, lat)?;
                stmt.execute(params![
                    person_id.as_str(),
                    plan_id.as_str(),
                    plan_index,
                    activity_index as i64,
                    lng,
                    lat
                ])
                .map_err(sqlite_err)?;
            }
        }
    }

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;

    Ok(PersonPlanPayload {
        plan_id,
        plan_index,
        selected,
        plan_blob: new_blob,
    })
}

#[tauri::command]
pub fn create_person(
    source_name: String,
    lng: f64,
    lat: f64,
    manager: State<'_, SessionManager>,
) -> Result<PersonPayload, EzError> {
    let path = population_db_path(&manager, &source_name)?;
    let mut connection = Connection::open(path).map_err(io_err)?;
    let tx = connection.transaction().map_err(sqlite_err)?;
    let crs = current_crs(&manager)?;
    let projector = Projector::new(&crs)?;
    let native = projector.unproject_lng_lat(lng, lat)?;

    let person_id = generate_person_id(&tx)?;
    let plan = PersonPlanPayload {
        plan_id: "p0".into(),
        plan_index: 0,
        selected: 1,
        plan_blob: default_plan_blob(1, native.x, native.y),
    };

    tx.execute(
        "INSERT INTO person_attributes (person_id, attributes_blob) VALUES (?1, NULL)",
        params![person_id.clone()],
    )
    .map_err(sqlite_err)?;
    tx.execute(
        "INSERT INTO person_plans (person_id, plan_id, plan_index, selected, plan_blob)
         VALUES (?1, ?2, ?3, ?4, ?5)",
        params![
            person_id.clone(),
            plan.plan_id.clone(),
            plan.plan_index,
            plan.selected,
            plan.plan_blob.clone()
        ],
    )
    .map_err(sqlite_err)?;
    tx.execute(
        "INSERT INTO plan_activity_coords
         (person_id, plan_id, plan_index, activity_index, lng, lat)
         VALUES (?1, ?2, ?3, 0, ?4, ?5)",
        params![person_id.clone(), plan.plan_id.clone(), plan.plan_index, lng, lat],
    )
    .map_err(sqlite_err)?;

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;

    Ok(PersonPayload {
        person_id,
        attributes_blob: None,
        plans: vec![plan],
    })
}

#[tauri::command]
pub fn create_plan(
    source_name: String,
    person_id: String,
    lng: f64,
    lat: f64,
    manager: State<'_, SessionManager>,
) -> Result<PersonPlanPayload, EzError> {
    let path = population_db_path(&manager, &source_name)?;
    let mut connection = Connection::open(path).map_err(io_err)?;
    let tx = connection.transaction().map_err(sqlite_err)?;
    let crs = current_crs(&manager)?;
    let projector = Projector::new(&crs)?;
    let native = projector.unproject_lng_lat(lng, lat)?;

    let next_index = tx
        .query_row(
            "SELECT COALESCE(MAX(plan_index) + 1, 0)
             FROM person_plans
             WHERE person_id = ?1",
            params![person_id.clone()],
            |row| row.get::<_, i64>(0),
        )
        .map_err(sqlite_err)?;

    let plan = PersonPlanPayload {
        plan_id: format!("p{next_index}"),
        plan_index: next_index,
        selected: 0,
        plan_blob: default_plan_blob(0, native.x, native.y),
    };

    tx.execute(
        "INSERT INTO person_plans (person_id, plan_id, plan_index, selected, plan_blob)
         VALUES (?1, ?2, ?3, ?4, ?5)",
        params![
            person_id.clone(),
            plan.plan_id.clone(),
            plan.plan_index,
            plan.selected,
            plan.plan_blob.clone()
        ],
    )
    .map_err(sqlite_err)?;
    tx.execute(
        "INSERT INTO plan_activity_coords
         (person_id, plan_id, plan_index, activity_index, lng, lat)
         VALUES (?1, ?2, ?3, 0, ?4, ?5)",
        params![person_id, plan.plan_id.clone(), plan.plan_index, lng, lat],
    )
    .map_err(sqlite_err)?;

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(plan)
}

#[tauri::command]
pub fn delete_plan(
    source_name: String,
    person_id: String,
    plan_id: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let path = population_db_path(&manager, &source_name)?;
    let mut connection = Connection::open(path).map_err(io_err)?;
    let tx = connection.transaction().map_err(sqlite_err)?;

    let count = tx
        .query_row(
            "SELECT COUNT(*) FROM person_plans WHERE person_id = ?1",
            params![person_id.clone()],
            |row| row.get::<_, i64>(0),
        )
        .map_err(sqlite_err)?;

    if count <= 1 {
        return Err(EzError::PopulationLastPlan { person_id });
    }

    let was_selected = tx
        .query_row(
            "SELECT selected FROM person_plans WHERE person_id = ?1 AND plan_id = ?2",
            params![person_id.clone(), plan_id.clone()],
            |row| row.get::<_, i64>(0),
        )
        .map_err(|err| match err {
            rusqlite::Error::QueryReturnedNoRows => EzError::PopulationPlanNotFound {
                person_id: person_id.clone(),
                plan_id: plan_id.clone(),
            },
            other => sqlite_err(other),
        })?;

    tx.execute(
        "DELETE FROM plan_activity_coords WHERE person_id = ?1 AND plan_id = ?2",
        params![person_id.clone(), plan_id.clone()],
    )
    .map_err(sqlite_err)?;
    tx.execute(
        "DELETE FROM person_plans WHERE person_id = ?1 AND plan_id = ?2",
        params![person_id.clone(), plan_id.clone()],
    )
    .map_err(sqlite_err)?;

    if was_selected == 1 {
        let replacement_plan_id = tx
            .query_row(
                "SELECT plan_id FROM person_plans WHERE person_id = ?1 ORDER BY plan_index ASC LIMIT 1",
                params![person_id.clone()],
                |row| row.get::<_, String>(0),
            )
            .map_err(sqlite_err)?;
        tx.execute(
            "UPDATE person_plans
             SET selected = CASE WHEN plan_id = ?1 THEN 1 ELSE 0 END
             WHERE person_id = ?2",
            params![replacement_plan_id, person_id.clone()],
        )
        .map_err(sqlite_err)?;
    }

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}

#[tauri::command]
pub fn update_plan_blob(
    source_name: String,
    person_id: String,
    plan_id: String,
    new_blob: String,
    new_coords: Vec<super::types::UpdatePlanCoordInput>,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let path = population_db_path(&manager, &source_name)?;
    let mut connection = Connection::open(path).map_err(io_err)?;
    let tx = connection.transaction().map_err(sqlite_err)?;

    let plan_index = tx
        .query_row(
            "SELECT plan_index
             FROM person_plans
             WHERE person_id = ?1 AND plan_id = ?2",
            params![person_id.clone(), plan_id.clone()],
            |row| row.get::<_, i64>(0),
        )
        .map_err(|err| match err {
            rusqlite::Error::QueryReturnedNoRows => EzError::PopulationPlanNotFound {
                person_id: person_id.clone(),
                plan_id: plan_id.clone(),
            },
            other => sqlite_err(other),
        })?;

    tx.execute(
        "UPDATE person_plans
         SET plan_blob = ?1
         WHERE person_id = ?2 AND plan_id = ?3",
        params![new_blob, person_id.clone(), plan_id.clone()],
    )
    .map_err(sqlite_err)?;

    tx.execute(
        "DELETE FROM plan_activity_coords WHERE person_id = ?1 AND plan_id = ?2",
        params![person_id.clone(), plan_id.clone()],
    )
    .map_err(sqlite_err)?;

    {
        let mut stmt = tx
            .prepare(
                "INSERT INTO plan_activity_coords
                 (person_id, plan_id, plan_index, activity_index, lng, lat)
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
            )
            .map_err(sqlite_err)?;
        for coord in new_coords {
            stmt.execute(params![
                person_id.as_str(),
                plan_id.as_str(),
                plan_index,
                coord.activity_index,
                coord.lng,
                coord.lat
            ])
            .map_err(sqlite_err)?;
        }
    }

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}

#[tauri::command]
pub fn delete_person(
    source_name: String,
    person_id: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let path = population_db_path(&manager, &source_name)?;
    let mut connection = Connection::open(path).map_err(io_err)?;
    let tx = connection.transaction().map_err(sqlite_err)?;

    tx.execute(
        "DELETE FROM plan_activity_coords WHERE person_id = ?1",
        params![person_id.clone()],
    )
    .map_err(sqlite_err)?;
    tx.execute(
        "DELETE FROM person_plans WHERE person_id = ?1",
        params![person_id.clone()],
    )
    .map_err(sqlite_err)?;
    let affected = tx
        .execute(
            "DELETE FROM person_attributes WHERE person_id = ?1",
            params![person_id.clone()],
        )
        .map_err(sqlite_err)?;

    if affected == 0 {
        return Err(EzError::PopulationPersonNotFound { person_id });
    }

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}

#[tauri::command]
pub fn set_plan_selected(
    source_name: String,
    person_id: String,
    plan_id: String,
    manager: State<'_, SessionManager>,
) -> Result<(), EzError> {
    let path = population_db_path(&manager, &source_name)?;
    let mut connection = Connection::open(path).map_err(io_err)?;
    let tx = connection.transaction().map_err(sqlite_err)?;

    let affected = tx
        .execute(
            "UPDATE person_plans
             SET selected = CASE WHEN plan_id = ?1 THEN 1 ELSE 0 END
             WHERE person_id = ?2",
            params![plan_id.clone(), person_id.clone()],
        )
        .map_err(sqlite_err)?;

    if affected == 0 {
        return Err(EzError::PopulationPlanNotFound { person_id, plan_id });
    }

    tx.commit().map_err(sqlite_err)?;
    mark_dirty(&manager)?;
    Ok(())
}

fn population_db_path(
    manager: &State<'_, SessionManager>,
    source_name: &str,
) -> Result<std::path::PathBuf, EzError> {
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
    Ok(path)
}

fn mark_dirty(manager: &State<'_, SessionManager>) -> Result<(), EzError> {
    let mut guard = manager
        .session
        .lock()
        .map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_mut().ok_or(EzError::NoActiveSession)?;
    session.dirty = true;
    Ok(())
}

fn generate_person_id(tx: &Transaction) -> Result<String, EzError> {
    let mut rng = rand::rng();
    loop {
        let id: u32 = rng.random_range(10_000_000..100_000_000);
        let candidate = id.to_string();
        let exists: bool = tx
            .query_row(
                "SELECT EXISTS(SELECT 1 FROM person_attributes WHERE person_id = ?1)",
                params![candidate],
                |row| row.get(0),
            )
            .map_err(sqlite_err)?;
        if !exists {
            return Ok(candidate);
        }
    }
}

fn default_plan_blob(selected: i64, x: f64, y: f64) -> String {
    format!(
        "<plan selected=\"{}\">\n  <activity type=\"home\" x=\"{:.6}\" y=\"{:.6}\"/>\n</plan>",
        if selected == 1 { "yes" } else { "no" },
        x,
        y
    )
}

fn io_err(err: rusqlite::Error) -> EzError {
    EzError::Io {
        message: format!("Failed to open population SQLite database: {err}"),
    }
}

fn sqlite_err(err: rusqlite::Error) -> EzError {
    EzError::Io {
        message: format!("Population edit failed: {err}"),
    }
}
