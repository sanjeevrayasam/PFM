from fastapi import FastAPI, Depends, HTTPException,status
from sqlalchemy.orm import Session
from app.models.database import Base, engine, database, get_db
from app.schema import auth
from app.auth import auth as authHandler

app = FastAPI()

def create_tables():
    Base.metadata.create_all(bind=engine)

@app.on_event("startup")
async def startup():
    # Connect to the database
    await database.connect()
    # Create tables
    create_tables()
    print("All tables created......")

@app.on_event("shutdown")
async def shutdown():
    await database.disconnect()

@app.post("/register", response_model=auth.User)
def register(user: auth.UserCreate, db: Session = Depends(get_db)):
    db_user = authHandler.get_user_by_username(db, user_name=user.username)
    if db_user:
        raise HTTPException(status_code=400, detail="Username already registered")
    return auth.create_user(db=db, user=user)


@app.post("/login")
def login(form_data: auth.Login, db: Session = Depends(get_db)):
    user = authHandler.authenticate_user(db, form_data.username, form_data.password)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect username or password",
            headers={"WWW-Authenticate": "Bearer"},
        )
    access_token, refresh_token = auth.create_tokens(user.id)
    return {"access_token": access_token, "refresh_token": refresh_token}


@app.post("/refresh")
def refresh_token(token: auth.TokenRefresh, db: Session = Depends(get_db)):
    user_id = authHandler.verify_token(token.refresh_token)
    if user_id is None:
        raise HTTPException(status_code=401, detail="Invalid token")
    access_token, refresh_token = auth.create_tokens(user_id)
    return {"access_token": access_token, "refresh_token": refresh_token}

